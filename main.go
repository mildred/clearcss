package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/css/scanner"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func main() {
	flag.Parse()

	err := TransformCSS(flag.Arg(0), os.Stdout, NewRules())
	if err != nil {
		log.Fatal(err)
	}
}

type Rules struct {
	rules map[string][]string
}

func NewRules() *Rules {
	return &Rules{
		rules: map[string][]string{},
	}
}

func TransformCSS(name string, out io.Writer, rules *Rules) error {
	in, err := os.Open(name)
	if err != nil {
		return err
	}
	defer in.Close()

	inputcss, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}

	scan := scanner.New(string(inputcss))
	trans := &transformer{scan, name, rules, NewRules(), nil}

	err = trans.processAny(out, true)
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}

type transformer struct {
	scan          *scanner.Scanner
	name          string
	rules         *Rules
	requiredRules *Rules
	props         []string
}

func (tr *transformer) next() (*scanner.Token, error) {
	tk := tr.scan.Next()
	if tk.Type == scanner.TokenEOF {
		return tk, io.EOF
	} else if tk.Type == scanner.TokenError {
		return tk, fmt.Errorf("%s: %s", tr.name, tk.String())
	}
	return tk, nil
}

func processSpace(s *scanner.Scanner) *scanner.Token {
	tk := s.Next()
	for tk.Type == scanner.TokenS {
		tk = s.Next()
	}
	return tk
}

func (tr *transformer) processAny(out io.Writer, root bool) error {
	stop := false
	var space string
	for !stop {
		tk, err := tr.next()
		if err != nil {
			return err
		}

		if tk.Type == scanner.TokenS {
			space = tk.Value
		}

		if tk.Type == scanner.TokenAtKeyword && tk.Value == "@require" {
			err = tr.processRequire()
			if err != nil {
				return err
			}
			continue
		}

		if tk.Type == scanner.TokenAtKeyword && tk.Value == "@extend" {
			err = tr.processExtend(space, out)
			if err != nil {
				return err
			}
			continue
		}

		stop = tk.Type == scanner.TokenChar && tk.Value == "}" && !root

		if !stop && tk.Type == scanner.TokenAtKeyword {
			tr.processDirective(tk, out)
		} else if !stop && tk.Type != scanner.TokenS {
			tr.processRule(tk, out)
		} else if out != nil {
			_, err = out.Write([]byte(tk.Value))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (tr *transformer) processRule(tk *scanner.Token, out io.Writer) error {
	var err error
	var rule string
	var rules []string
	var property string
	for {
		if out != nil {
			_, err = out.Write([]byte(tk.Value))
			if err != nil {
				return err
			}
		}

		if tk.Type == scanner.TokenChar && tk.Value == ";" {
			property += tk.Value
			tr.props = append(tr.props, property)
			return nil
		} else if tk.Type == scanner.TokenChar && tk.Value == "{" {
			tr.props = nil
			rules = append(rules, rule)
			err = tr.processAny(out, false)
			for _, r := range rules {
				tr.rules.rules[r] = append(tr.rules.rules[r], tr.props...)
			}
			return err
		} else {
			property += tk.Value
			if tk.Type == scanner.TokenChar && tk.Value == "," {
				rules = append(rules, rule)
				rule = ""
			} else if tk.Type != scanner.TokenS {
				rule = rule + tk.Value
			}
		}

		tk, err = tr.next()
		if err != nil {
			return err
		}
	}
}

func (tr *transformer) processDirective(tk *scanner.Token, out io.Writer) error {
	var err error
	for {
		if out != nil {
			_, err = out.Write([]byte(tk.Value))
			if err != nil {
				return err
			}
		}

		if tk.Type == scanner.TokenChar && tk.Value == ";" {
			return nil
		} else if tk.Type == scanner.TokenChar && tk.Value == "{" {
			return tr.processAny(out, false)
		}

		tk, err = tr.next()
		if err != nil {
			return err
		}
	}
}

func (tr *transformer) processRequire() error {
	tk := processSpace(tr.scan)
	for tk.Type != scanner.TokenChar || tk.Value != ";" {
		path := filepath.Join(filepath.Dir(tr.name), tk.Value[1:len(tk.Value)-1])
		err := TransformCSS(path, nil, tr.requiredRules)
		if err != nil {
			return err
		}
		tk = processSpace(tr.scan)
	}
	return nil
}

func (tr *transformer) processExtend(indent string, out io.Writer) error {
	var rule string
	var rules []string
	var err error
	tk := processSpace(tr.scan)
	for tk.Type != scanner.TokenChar || tk.Value != ";" {
		if tk.Type == scanner.TokenChar && tk.Value == "," {
			rules = append(rules, rule)
			rule = ""
		} else {
			rule += tk.Value
		}
		tk = processSpace(tr.scan)
	}
	if rule != "" {
		rules = append(rules, rule)
	}
	if out == nil {
		return nil
	}

	for i, rule := range rules {
		if i != 0 && out != nil {
			fmt.Fprint(out, indent)
		}
		var prefix = ""
		if len(rules) > 1 {
			prefix = fmt.Sprintf("[%d]", i)
		}
		if properties, ok := tr.requiredRules.rules[rule]; ok {
			_, err = fmt.Fprintf(out, "/* @extend%s %s */\n", prefix, rule)
			if err != nil {
				return err
			}
			for _, prop := range properties {
				_, err = fmt.Fprintf(out, "%s  %s", indent, prop)
				if err != nil {
					return err
				}
			}
			_, err = fmt.Fprint(out, "\n")
			if err != nil {
				return err
			}
		} else {
			_, err = fmt.Fprintf(out, "/* NOT FOUND @extend%s %s */\n", prefix, rule)
		}
	}

	return nil
}
