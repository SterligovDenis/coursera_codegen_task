package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
	"text/template"
)

// код писать тут

type RecvStruct struct {
	TypeName string
	Methods  []Method
}

type Method struct {
	URL      string
	Auth     bool
	Method   string
	Name     string
	RecvType string
	Params   []ApiStruct
}

type ApiStruct struct {
	Name   string
	Fields []FieldValidator
}

type FieldValidator struct {
	FieldName    string
	FieldType    string
	Alias        string
	DefaultValue string
	Enum         []string
	Required     bool
	IsMin        bool
	IsMax        bool
	Min          int
	Max          int
}

var handlerTpl = template.Must(template.New("intTpl").Parse(`
	{{$recv := .TypeName}}
	func (this *{{$recv}}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
		{{range .Methods}}
			if r.URL.Path == "{{.URL}}" {
				this.handler{{.Name}}(w, r)
				return
			}
		{{end}}
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{\"error\": \"unknown method\"}"))
	}

	{{range .Methods}}
		func (this *{{$recv}}) handler{{.Name}}(w http.ResponseWriter, r *http.Request) {
			{{if ne .Method ""}}
				if r.Method != "{{.Method}}" {
					w.WriteHeader(http.StatusNotAcceptable)
					w.Write([]byte("{\"error\": \"bad method\"}"))
					return
				}
			{{end}}
	
			if r.Method == http.MethodPost {
				r.ParseForm()
			}
			
			{{if eq .Auth true}}
				if r.Header.Get("X-Auth") != "100500" {
					w.WriteHeader(http.StatusForbidden)
					w.Write([]byte("{\"error\": \"unauthorized\"}"))
					return
				}
			{{end}}
			{{range .Params}}
				{{$p := len .Fields}}
				{{if ne $p 0}}
					v := {{.Name}}{}
					{{range .Fields}}
						{{if eq .FieldType "int"}}
							var val int
							var err error
							if r.Method == http.MethodPost {
								val, err = strconv.Atoi(r.FormValue("{{.Alias}}"))
							} else {
								val, err = strconv.Atoi(r.URL.Query().Get("{{.Alias}}"))
							}
							if err != nil {
								w.WriteHeader(http.StatusBadRequest)
								w.Write([]byte("{\"error\": \"{{.Alias}} must be int\"}"))
								return
							}
							v.{{.FieldName}} = val
						{{else}}
							if r.Method == http.MethodPost {
								v.{{.FieldName}} = r.FormValue("{{.Alias}}")
							} else {
								v.{{.FieldName}} = r.URL.Query().Get("{{.Alias}}")
							}
						{{end}}

						{{if eq .FieldType "string"}}
							{{if .Required}}
								if v.{{.FieldName}} == "" {
									w.WriteHeader(http.StatusBadRequest)
									w.Write([]byte("{\"error\": \"{{.Alias}} must me not empty\"}"))
									return
								}
							{{end}}
							{{if ne .DefaultValue ""}}
								if v.{{.FieldName}} == "" {
									v.{{.FieldName}} = "{{.DefaultValue}}"
								}
							{{end}}
						{{else}}
							{{if eq .Required true}}
								if v.{{.FieldName}} == 0 {
									w.WriteHeader(http.StatusBadRequest)
									w.Write([]byte("{\"error\": \"{{.Alias}} must me not empty\"}"))
									return
								}
							{{end}}
							{{if ne .DefaultValue ""}}
								if v.{{.FieldName}} == 0 {
									v.{{.FieldName}} = {{.DefaultValue}}
								}
							{{end}}
						{{end}}

						{{$l := len .Enum}}
						{{if ne $l 0}}
							inEnum := false
							var errArr []string
							{{$fieldName := .FieldName}}
							{{$alias := .Alias}}
							{{range .Enum}}
							 	errArr = append(errArr, "{{.}}")
								if v.{{$fieldName}} == "{{.}}" {
									inEnum = true
								}
							{{end}}	
							if !inEnum {
								w.WriteHeader(http.StatusBadRequest)
								w.Write([]byte("{\"error\": \"{{$alias}} must be one of ["+strings.Join(errArr, ", ")+"]\"}"))
								return
							}
						{{end}}

						{{if .IsMin}}
							{{if eq .FieldType "string"}}
								if len(v.{{.FieldName}}) < {{.Min}} {
									w.WriteHeader(http.StatusBadRequest)
									w.Write([]byte("{\"error\": \"{{.Alias}} len must be >= {{.Min}}\"}"))
									return
								}
							{{else}}
								if v.{{.FieldName}} < {{.Min}} {
									w.WriteHeader(http.StatusBadRequest)
									w.Write([]byte("{\"error\": \"{{.Alias}} must be >= {{.Min}}\"}"))
									return
								}
							{{end}}
						{{end}}

						{{if .IsMax}}
							{{if eq .FieldType "string"}}
								if len(v.{{.FieldName}}) > {{.Max}} {
									w.WriteHeader(http.StatusBadRequest)
									w.Write([]byte("{\"error\": \"{{.Alias}} len must be <= {{.Max}}\"}"))
									return
								}
							{{else}}
								if v.{{.FieldName}} > {{.Max}} {
									w.WriteHeader(http.StatusBadRequest)
									w.Write([]byte("{\"error\": \"{{.Alias}} must be <= {{.Max}}\"}"))
									return
								}
							{{end}}
						{{end}}
						
					{{end}}
				{{end}}
			{{end}}
			ctx := r.Context()
			res, err := this.{{.Name}}(ctx, v)
			if err != nil {
				if apiErr, ok := err.(ApiError); ok {
					w.WriteHeader(apiErr.HTTPStatus)
					w.Write([]byte("{\"error\": \"" + apiErr.Error() + "\"}"))
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("{\"error\": \"" + err.Error() + "\"}"))
				}
				return
			}
			resMap := make(map[string]interface{})
			resMap["error"] = interface{}("")
			resMap["response"] = interface{}(res)
			jsonStr, err := json.Marshal(resMap)
			w.Write(jsonStr)	
		}
	{{end}}
`))

func parseValidatorStr(f *ast.Field) FieldValidator {
	validatorStr := f.Tag.Value
	if validatorStr == "" {
		return FieldValidator{}
	}
	pos := strings.Index(validatorStr, `"`) + 1
	validatorStr = validatorStr[pos : len(validatorStr)-2]
	validators := strings.Split(validatorStr, ",")
	fvalidator := FieldValidator{}
	var err error
	for _, v := range validators {
		v = strings.Trim(v, " ")
		if strings.Contains(v, "required") {
			fvalidator.Required = true
		} else if strings.Contains(v, "min=") {
			fvalidator.IsMin = true
			str := strings.Split(v, "=")
			fvalidator.Min, err = strconv.Atoi(strings.Trim(str[1], " "))
			if err != nil {
				panic("Error with convertation! " + err.Error())
			}
		} else if strings.Contains(v, "max=") {
			fvalidator.IsMax = true
			str := strings.Split(v, "=")
			fvalidator.Max, err = strconv.Atoi(strings.Trim(str[1], " "))
			if err != nil {
				panic("Error with convertation! " + err.Error())
			}
		} else if strings.Contains(v, "paramname=") {
			str := strings.Split(v, "=")
			fvalidator.Alias = str[1]
		} else if strings.Contains(v, "default=") {
			str := strings.Split(v, "=")
			fvalidator.DefaultValue = str[1]
		} else if strings.Contains(v, "enum=") {
			str := strings.Split(v, "=")
			fvalidator.Enum = strings.Split(str[1], "|")
		}
	}
	xv := f.Type.(*ast.Ident)
	fvalidator.FieldType = xv.Name
	fvalidator.FieldName = f.Names[0].Name
	if fvalidator.Alias == "" {
		fvalidator.Alias = strings.ToLower(f.Names[0].Name)
	}
	return fvalidator
}

func getFuncParams(f *ast.FuncDecl) []string {
	if f.Type.Params != nil {
		params := make([]string, 0)
		for _, l := range f.Type.Params.List {
			switch xv := l.Type.(type) {
			case *ast.SelectorExpr:
				t := xv.Sel.Name
				if p, ok := xv.X.(*ast.Ident); ok {
					params = append(params, p.Name+"."+t)
				}
			case *ast.Ident:
				params = append(params, xv.Name)
			}
		}
		return params
	}
	return nil
}

func genMethodsAndValidators(node *ast.File) []Method {
	methods := make([]Method, 0)
	validators := make(map[string][]FieldValidator)
	funcParams := make(map[string]map[string][]string)
	for _, spec := range node.Decls { // функции
		f, _ := spec.(*ast.FuncDecl)
		if f != nil {
			com := f.Doc.Text()
			if strings.Contains(com, "apigen:api") {
				jsonComStr := strings.Replace(com, "apigen:api", "", -1)
				jsonComStr = strings.Trim(jsonComStr, " \n")
				curMethod := Method{}
				if err := json.Unmarshal([]byte(jsonComStr), &curMethod); err != nil {
					panic(err.Error())
				}
				var curStructType string
				switch xv := f.Recv.List[0].Type.(type) {
				case *ast.StarExpr:
					if si, ok := xv.X.(*ast.Ident); ok {
						curStructType = si.Name
					}
				case *ast.Ident:
					curStructType = xv.Name
				}
				curMethod.RecvType = curStructType
				curMethod.Name = f.Name.Name
				if _, ok := funcParams[curStructType]; !ok {
					funcParams[curStructType] = make(map[string][]string)
				}
				funcParams[curStructType][curMethod.Name] = getFuncParams(f)
				methods = append(methods, curMethod)
				// curMethod.Methods = append(methods[curStructType].Methods, curMethod)
			}
		} else { // структуры
			g, _ := spec.(*ast.GenDecl)
			for _, s := range g.Specs {
				curType, ok := s.(*ast.TypeSpec)
				if !ok {
					// fmt.Printf("SKIP %T is not ast.TypeSpec\n", s)
					continue
				}
				curStruct, ok := curType.Type.(*ast.StructType)
				if !ok {
					// fmt.Printf("SKIP %T is not ast.StructType\n", curType)
					continue
				}
				if curStruct.Fields != nil {
					for _, structField := range curStruct.Fields.List {
						if structField.Tag != nil && strings.Contains(structField.Tag.Value, "apivalidator") {
							validators[curType.Name.Name] = append(validators[curType.Name.Name], parseValidatorStr(structField))
						}
					}
				}
			}
		}
	}
	// после того как у нас есть вся информация о функциях и структурах
	// распихаем параметры функций по функциям
	for key, m := range methods {
		for _, funcParamType := range funcParams[m.RecvType][m.Name] {
			as := ApiStruct{}
			as.Name = funcParamType
			if _, ok := validators[funcParamType]; ok {
				as.Fields = validators[funcParamType]
			}
			methods[key].Params = append(methods[key].Params, as)
		}
	}
	return methods
}

func printImports(out *os.File) {
	fmt.Fprintln(out, `package main`)
	fmt.Fprintln(out, `import "encoding/json"`)
	fmt.Fprintln(out, `import "strconv"`)
	fmt.Fprintln(out, `import "strings"`)
	fmt.Fprintln(out, `import "net/http"`)
}

func genCode(out *os.File, fin string) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, fin, nil, parser.ParseComments)
	if err != nil {
		panic("Parsing input file error! " + err.Error())
	}
	methods := genMethodsAndValidators(node)
	printImports(out)
	recvs := make(map[string]*RecvStruct)
	for _, m := range methods {
		if _, ok := recvs[m.RecvType]; !ok {
			recvs[m.RecvType] = &RecvStruct{}
		}
		recvs[m.RecvType].TypeName = m.RecvType
		recvs[m.RecvType].Methods = append(recvs[m.RecvType].Methods, m)
	}
	for _, s := range recvs {
		handlerTpl.Execute(out, s)
	}
}

func main() {
	args := os.Args[1:]
	if len(args) < 2 {
		panic("You must specify two arguments!")
	}
	out, err := os.Create(args[1])
	if err != nil {
		panic("Creating file error!")
	}
	defer out.Close()
	genCode(out, args[0])
}
