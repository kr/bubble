package parser

import (
	"fmt"
	"go/scanner"
	"go/token"
	"io/ioutil"
	"os"
	"strings"

	"github.com/kr/bubble/ast"
)

type Mode int

const (
	Debug Mode = 1 << iota
)

type parser struct {
	mode    Mode
	fileSet *token.FileSet
	scanner scanner.Scanner

	pos token.Pos
	tok token.Token
	lit string
}

func (p *parser) next() {
	p.pos, p.tok, p.lit = p.scanner.Scan()
	if p.mode&Debug != 0 {
		pos := p.fileSet.Position(p.pos)
		fmt.Println(pos, "\ttoken", p.tok, p.lit)
	}
}

func (p *parser) want(tok token.Token) {
	if p.tok != tok {
		// TODO(kr): don't crash here
		p.errorf("error tok = %v want %v", p.tok, tok)
	}
	p.next()
}

func handleError(pos token.Position, msg string) {
	fmt.Fprintln(os.Stderr, "error", pos, msg)
}

func Parse(fset *token.FileSet, mode Mode) (*ast.Package, error) {
	var p parser
	p.mode = mode
	p.fileSet = fset
	pkg := new(ast.Package)
	var err error
	fset.Iterate(func(f *token.File) bool {
		var file *ast.File
		file, err = p.parseFile(f)
		if err != nil {
			return false
		}
		pkg.Files = append(pkg.Files, file)
		return true
	})
	if err != nil {
		return nil, err
	}
	return pkg, nil
}

func (p *parser) parseFile(f *token.File) (*ast.File, error) {
	text, err := readSource(f.Name())
	if err != nil {
		return nil, err
	}
	p.scanner.Init(f, text, handleError, 0)
	file := new(ast.File)
	p.next()
	for p.tok == token.IMPORT {
		imps := p.parseImportStmt()
		file.Imports = append(file.Imports, imps...)
		p.want(token.SEMICOLON)
	}
	for {
		switch p.tok {
		case token.FUNC:
			x, err := p.parseFuncDecl()
			if err != nil {
				return nil, err
			}
			file.Funcs = append(file.Funcs, x)
			p.want(token.SEMICOLON)
		case token.EOF:
			return file, nil
		case token.IMPORT:
			p.errorf("import after declaration")
		default:
			// TODO(kr): don't crash here
			p.errorf("unexpected: %v", p.tok)
		}
	}
}

func (p *parser) parseImportStmt() []*ast.ImportSpec {
	p.want(token.IMPORT)
	// TODO(kr): imports grouped with parentheses
	if p.tok != token.STRING {
		p.errorf("error tok = %v want string", p.tok)
	}
	defer p.next()
	return []*ast.ImportSpec{
		{Path: &ast.BasicLit{p.tok, p.lit}},
	}
}

func (p *parser) parseFuncDecl() (*ast.FuncDecl, error) {
	p.want(token.FUNC)
	name := &ast.Ident{p.lit}
	p.want(token.IDENT)
	p.want(token.LPAREN)
	var params []*ast.Ident
	if p.tok != token.RPAREN {
		params = p.parseVarList()
	}
	p.want(token.RPAREN)
	body := p.parseBlockStmt()
	return &ast.FuncDecl{Name: name, Params: params, Body: body}, nil
}

func (p *parser) parseVarList() []*ast.Ident {
	var vars []*ast.Ident
	for p.tok == token.IDENT {
		vars = append(vars, &ast.Ident{p.lit})
		p.next()
		if p.tok != token.COMMA {
			break
		}
		p.next()
	}
	return vars
}

func (p *parser) parseBlockStmt() *ast.BlockStmt {
	p.want(token.LBRACE)
	body := new(ast.BlockStmt)
	for p.tok != token.RBRACE {
		stmt := p.parseStmt()
		body.List = append(body.List, stmt)
	}
	p.want(token.RBRACE)
	return body
}

func (p *parser) parseStmt() ast.Stmt {
	defer p.want(token.SEMICOLON)
	switch p.tok {
	case token.IF:
		return p.parseIf()
	case token.RETURN:
		return p.parseReturn()
	default:
		return p.parseExprStmt()
	}
}

func (p *parser) parseIf() ast.Stmt {
	p.want(token.IF)
	cond := p.parseExpr()
	body := p.parseBlockStmt()
	s := &ast.IfStmt{Cond: cond, Body: body}
	if p.tok == token.ELSE {
		p.next()
		switch p.tok {
		case token.IF:
			s.Else = p.parseIf()
		case token.LBRACE:
			s.Else = p.parseBlockStmt()
		default:
			p.errorf("expected if or {")
		}
	}
	return s
}

func (p *parser) parseReturn() *ast.ReturnStmt {
	p.want(token.RETURN)
	x := p.parseExpr()
	return &ast.ReturnStmt{x}
}

func (p *parser) parseExprStmt() ast.Stmt {
	x := p.parseExpr()
	switch p.tok {
	case token.DEFINE, token.ASSIGN:
		tok := p.tok
		p.next()
		y := p.parseExpr()
		return &ast.AssignStmt{Lhs: x, Tok: tok, Rhs: y}
	}
	return &ast.ExprStmt{x}
}

func (p *parser) parseExpr() ast.Expr {
	e := p.parseTerm()
	for p.tok == token.ADD || p.tok == token.SUB {
		t := p.tok
		p.next()
		e = &ast.BinaryExpr{X: e, Op: t, Y: p.parseTerm()}
	}
	return e
}

func (p *parser) parseTerm() ast.Expr {
	e := p.parseCall()
	for p.tok == token.MUL || p.tok == token.QUO {
		t := p.tok
		p.next()
		e = &ast.BinaryExpr{X: e, Op: t, Y: p.parseCall()}
	}
	return e
}

func (p *parser) parseCall() ast.Expr {
	x := p.parseSel()
	if p.tok != token.LPAREN {
		return x
	}
	call := &ast.CallExpr{Fun: x}
	p.want(token.LPAREN)
	for p.tok != token.RPAREN {
		call.Args = append(call.Args, p.parseExpr())
		if p.tok == token.RPAREN {
			break
		}
		p.want(token.COMMA)
	}
	p.want(token.RPAREN)
	return call
}

func (p *parser) parseSel() ast.Expr {
	x := p.parseAtom()
	for p.tok == token.PERIOD {
		p.next()
		lit := p.lit
		p.want(token.IDENT)
		x = &ast.SelectorExpr{X: x, Sel: &ast.Ident{lit}}
	}
	return x
}

func (p *parser) parseAtom() ast.Expr {
	switch tok, lit := p.tok, p.lit; tok {
	case token.IDENT:
		p.next()
		return &ast.Ident{lit}
	case token.INT, token.STRING:
		p.next()
		return &ast.BasicLit{tok, lit}
	case token.FUNC:
		return p.parseFuncLit()
	case token.AND:
		p.next()
		body := p.parseExpr()
		return &ast.ShortFuncLit{Body: body}
	case token.LPAREN:
		p.next()
		defer p.want(token.RPAREN)
		return p.parseExpr()
	}
	// TODO(kr): don't crash here
	p.errorf("error tok = %v want ident or literal", p.tok)
	return nil
}

func (p *parser) parseFuncLit() ast.Expr {
	p.want(token.FUNC)
	p.want(token.LPAREN)
	var params []*ast.Ident
	if p.tok != token.RPAREN {
		params = p.parseVarList()
	}
	p.want(token.RPAREN)
	body := p.parseBlockStmt()
	return &ast.FuncLit{Params: params, Body: body}
}

// errorf prints the current position p.pos followed by
// a formatted error message.
func (p *parser) errorf(format string, v ...interface{}) {
	s := p.fileSet.Position(p.pos)
	v = append([]interface{}{s}, v...)
	format = strings.TrimSpace(format) + "\n"
	fmt.Fprintf(os.Stderr, "%s: "+format, v...)
	os.Exit(1)
}

func readSource(name string) ([]byte, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return b, nil
}
