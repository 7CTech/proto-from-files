package main

import (
	"reflect"
	"strings"
	"go/parser"
	"go/token"
	"go/ast"
	"strconv"
	"os"
)

/*
Utility file to generate proto files. It takes a service name, then list of files as arguments, space seperated. all public methods in these files
are put into the proto file

This seems stupid, cause its doing the exact opposite of normal protobuf, but it has come in handy
 */
func main() {
	println(genProto(os.Args))
}

func genProto(args []string) string {
	sb := strings.Builder{}

	funcDecls := make([]*ast.FuncDecl, 0)

	for i, element := range args {
		if (i == 0 || i == 1) {
			continue;
		}
		funcDecls = append(funcDecls, getFuncDecls(element)...)
	}

	sb.WriteString("syntax = \"proto3\";\n\n")


	sb.WriteString("import \"google/protobuf/empty.proto\";\n\n")
	sb.WriteString("service ")
	sb.WriteString(args[1])
	sb.WriteString(" {\n")
	if len(funcDecls) != 0 {
		for _, funcDecl := range funcDecls {
			sb.WriteString(parseDeclaration(funcDecl))
		}
	}
	sb.WriteString("}")

	if len(funcDecls) != 0 {
		for _, funcDecl := range funcDecls {
			sb.WriteString(parseParamsAndResults(funcDecl))
		}
	}

	return sb.String()
}

func getFuncDecls(fileName string) []*ast.FuncDecl {
	newDecls := make([]*ast.FuncDecl, 0)
	fs := token.NewFileSet()

	f, err := parser.ParseFile(fs, fileName, nil, parser.AllErrors)
	if err == nil {
		for _, element := range f.Decls {
			if reflect.TypeOf(element) == reflect.TypeOf((*ast.FuncDecl)(nil)) {
				newDecls = append(newDecls, element.(*ast.FuncDecl))
			}
		}
	}
	return newDecls
}

func convertToProtoType(goType string) string {
	if goType == "float64" {
		return "double"
	} else if goType == "float32" {
		return "float"
	} else if goType == "int" || goType == "int8" || goType == "int16" {
		return "int32"
	} else if goType == "uint" || goType == "uint8" || goType == "uint16" {
		return "uint32"
	}
	return goType
}

func getResultName(functionName string) string {
	var strippedString = &functionName
	if strings.HasPrefix(functionName, "get") {
		strippedString = &strings.SplitAfter(functionName, "get")[1]
	} else if strings.HasPrefix(functionName, "Get") {
		strippedString = &strings.SplitAfter(functionName, "Get")[1]
	}
	return strings.ToLower((*strippedString)[0:1]) + (*strippedString)[1:]
}

func parseDeclaration(function *ast.FuncDecl) string {
	sb := strings.Builder{}
	sb.WriteString("\t rpc ")
	sb.WriteString(function.Name.Name)
	sb.WriteString("(")
	if len(function.Type.Params.List) == 0 {
		sb.WriteString("google.protobuf.Empty")
	} else {
		sb.WriteString(function.Name.Name)
		sb.WriteString("Request")
	}
	sb.WriteString(") returns (")
	if len(function.Type.Results.List) == 0 {
		sb.WriteString("google.protobuf.Empty")
	} else {
		sb.WriteString(function.Name.Name)
		sb.WriteString("Reply")
	}
	sb.WriteString(") {}\n")
	return sb.String()
}

func parseParamsAndResults(function *ast.FuncDecl) string {
	sb := strings.Builder{}
	if function.Name.IsExported() {
		//1. params
		if len(function.Type.Params.List) > 0 {
			sb.WriteString("\n\nmessage ")
			sb.WriteString(function.Name.Name)
			sb.WriteString("Request")
			sb.WriteString(" {\n")
			for index, element := range function.Type.Params.List {
				sb.WriteString("\t")
				sb.WriteString(convertToProtoType(element.Type.(*ast.Ident).Name))
				sb.WriteString(" ")
				sb.WriteString(element.Names[0].Name)
				sb.WriteString(" = ")
				sb.WriteString(strconv.Itoa(index + 1))
				sb.WriteString(";\n")
			}
			sb.WriteString("}")
		}

		//2. return
		if len(function.Type.Results.List) > 0 {
			sb.WriteString("\n\nmessage ")
			sb.WriteString(function.Name.Name)
			sb.WriteString("Reply {\n")
			for index, element := range function.Type.Results.List {
				sb.WriteString("\t")
				if reflect.TypeOf(element.Type) == reflect.TypeOf((*ast.Ident)(nil)) {
					//primitive
					sb.WriteString(convertToProtoType(element.Type.(*ast.Ident).Name))
				} else if reflect.TypeOf(element.Type) == reflect.TypeOf((*ast.ArrayType)(nil)) {
					//array
					sb.WriteString("repeated ")
					sb.WriteString(convertToProtoType(element.Type.(*ast.ArrayType).Elt.(*ast.Ident).Name))
				}

				sb.WriteString(" ")

				sb.WriteString(getResultName(function.Name.Name))

				sb.WriteString(" = ")
				sb.WriteString(strconv.Itoa(index + 1))
				sb.WriteString(";\n")
			}
			sb.WriteString("}")
		}
	}
	return sb.String()
}
