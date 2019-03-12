package main

import (
	"encoding/xml"
	"fmt"
	"github.com/gswly/gomavlib"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

type MavlinkEnumEntry struct {
	Value       string `xml:"value,attr"`
	Name        string `xml:"name,attr"`
	Description string `xml:"description"`
}

type MavlinkEnum struct {
	Description string              `xml:"description"`
	Entries     []*MavlinkEnumEntry `xml:"entry"`
}

type MavlinkField struct {
	Extension   bool   `xml:"-"`
	Type        string `xml:"type,attr"`
	Name        string `xml:"name,attr"`
	Enum        string `xml:"enum,attr"`
	Description string `xml:",innerxml"`
}

type MavlinkMessage struct {
	Id          uint32
	Name        string
	Description string
	Fields      []*MavlinkField
}

// we must unmarshal manually due to extension fields
func (m *MavlinkMessage) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// unmarshal attributes
	for _, a := range start.Attr {
		switch a.Name.Local {
		case "id":
			v, _ := strconv.Atoi(a.Value)
			m.Id = uint32(v)
		case "name":
			m.Name = a.Value
		}
	}

	inExtensions := false
	for {
		t, _ := d.Token()
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			switch se.Name.Local {
			case "description":
				err := d.DecodeElement(&m.Description, &se)
				if err != nil {
					return err
				}

			case "extensions":
				inExtensions = true

			case "field":
				field := &MavlinkField{Extension: inExtensions}
				err := d.DecodeElement(&field, &se)
				if err != nil {
					return err
				}
				m.Fields = append(m.Fields, field)
			}
		}
	}
	return nil
}

type MavlinkDefinition struct {
	Name     string            `xml:"-"`
	Version  int               `xml:"version"`
	Dialect  int               `xml:"dialect"`
	Includes []string          `xml:"include"`
	Enums    []*MavlinkEnum    `xml:"enums>enum"`
	Messages []*MavlinkMessage `xml:"messages>message"`
}

var tpl = template.Must(template.New("").Parse(`
// autogenerated with dialgen. do not edit.

package {{ .Name }}

import (
	"github.com/gswly/gomavlib"
)
{{/**/}}
{{- range .Defs }}
// {{ .Name }}
{{/**/}}
{{- range .Messages }}
type Message{{ .Name }} struct {
{{- range .Fields }}
    {{ .Name }} {{ .Type }}
{{- end }}
}

func (*Message{{ .Name }}) GetId() uint32 {
    return {{ .Id }}
}
{{ end }}
{{- end }}

var Dialect = []gomavlib.Message{
{{- range .Defs }}
    // {{ .Name }}
{{- range .Messages }}
    &Message{{ .Name }}{},
{{- end }}
{{- end }}
}
`))

func fieldTypeToGo(f *MavlinkField) string {
	typ := f.Type

	arrayLen := ""
	tags := make(map[string]string)

	if typ == "uint8_t_mavlink_version" {
		typ = "uint8_t"
	}

	// string or array
	re := regexp.MustCompile("^(.+?)\\[([0-9]+)\\]$")
	if matches := re.FindStringSubmatch(typ); matches != nil {
		// string
		if matches[1] == "char" {
			tags["mavlen"] = matches[2]
			typ = "char"
			// array
		} else {
			arrayLen = matches[2]
			typ = matches[1]
		}
	}

	// extension
	if f.Extension == true {
		tags["mavext"] = "true"
	}

	typ = gomavlib.MsgTypeXmlToGo[typ]

	out := ""
	if arrayLen != "" {
		out += "[" + arrayLen + "]"
	}
	out += typ
	if len(tags) > 0 {
		var tmp []string
		for k, v := range tags {
			tmp = append(tmp, fmt.Sprintf("%s:\"%s\"", k, v))
		}
		out += " `" + strings.Join(tmp, ",") + "`"
	}
	return out
}

func main() {
	outfile := kingpin.Flag("output", "output file").Required().String()
	mainDefAddr := kingpin.Arg("xml", "a path or url pointing to a Mavlink dialect definition in XML format").Required().String()
	kingpin.CommandLine.Help = "Generate a Mavlink dialect library from a definition file.\n" +
		"Example: dialgen \\\n--output=dialect.go \\\nhttps://raw.githubusercontent.com/mavlink/mavlink/master/message_definitions/v1.0/common.xml"
	kingpin.Parse()

	err := do(*outfile, *mainDefAddr)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}

func do(outfile string, mainDefAddr string) error {
	if strings.HasSuffix(outfile, ".go") == false {
		return fmt.Errorf("output file must end with .go")
	}

	isRemote := func() bool {
		_, err := url.ParseRequestURI(mainDefAddr)
		return err == nil
	}()
	processed := make(map[string]struct{})
	defs := []MavlinkDefinition{}

	// parse all definitions recursively
	var process func(defAddr string) error
	process = func(defAddr string) error {
		// skip already processed
		if _, ok := processed[defAddr]; ok {
			return nil
		}
		processed[defAddr] = struct{}{}

		fmt.Printf("parsing %s...\n", defAddr)

		// get definition content
		content, err := func() ([]byte, error) {
			if isRemote == true {
				byt, err := dlUrl(defAddr)
				if err != nil {
					return nil, fmt.Errorf("unable to download: %s", err)
				}
				return byt, nil

			} else {
				byt, err := ioutil.ReadFile(defAddr)
				if err != nil {
					return nil, fmt.Errorf("unable to open: %s", err)
				}
				return byt, nil
			}
		}()
		if err != nil {
			return err
		}

		// parse definition
		var def MavlinkDefinition
		err = xml.Unmarshal(content, &def)
		if err != nil {
			return fmt.Errorf("unable to decode: %s", err)
		}

		addrPath, addrName := filepath.Split(defAddr)
		def.Name = addrName

		// process includes
		for _, inc := range def.Includes {
			// prepend url to remote address
			if isRemote == true {
				inc = addrPath + inc
			}
			err := process(inc)
			if err != nil {
				return err
			}
		}

		// convert strings to go format
		for _, msg := range def.Messages {
			msg.Name = gomavlib.MsgNameXmlToGo(msg.Name)
			for _, f := range msg.Fields {
				f.Name = gomavlib.MsgFieldXmlToGo(f.Name)
				f.Type = fieldTypeToGo(f)
			}
		}

		defs = append(defs, def)
		return nil
	}
	err := process(mainDefAddr)
	if err != nil {
		return err
	}

	// create output folder
	dir, _ := filepath.Split(outfile)
	os.Mkdir(dir, 0755)

	// open file
	f, err := os.Create(outfile)
	if err != nil {
		return err
	}
	defer f.Close()

	// dump
	tmp := map[string]interface{}{
		"Name": func() string {
			_, name := filepath.Split(mainDefAddr)
			return strings.TrimSuffix(name, ".xml")
		}(),
		"Defs": defs,
	}
	err = tpl.Execute(f, tmp)
	if err != nil {
		return err
	}
	return nil
}

func dlUrl(desturl string) ([]byte, error) {
	res, err := http.Get(desturl)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad return code: %v", res.StatusCode)
	}

	byt, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return byt, nil
}
