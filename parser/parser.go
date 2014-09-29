package parser

import (
	"fmt"
	"go/scanner"
	"go/token"
	"io/ioutil"
	"log"
	"os"

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
		pos := p.fileSet.Position(p.pos)
		// TODO(kr): don't crash here
		log.Fatalln(pos, "error tok =", p.tok, "want", tok)
	}
	p.next()
}

func handleError(pos token.Position, msg string) {
	fmt.Fprintln(os.Stderr, "error", pos, msg)
}

func Parse(fset *token.FileSet, mode Mode) (*ast.Program, error) {
	var p parser
	p.mode = mode
	p.fileSet = fset
	prog := new(ast.Program)
	var err error
	fset.Iterate(func(f *token.File) bool {
		var file *ast.File
		file, err = p.parseFile(f)
		if err != nil {
			return false
		}
		prog.Files = append(prog.Files, file)
		return true
	})
	if err != nil {
		return nil, err
	}
	return prog, nil
}

func (p *parser) parseFile(f *token.File) (*ast.File, error) {
	text, err := readSource(f.Name())
	if err != nil {
		return nil, err
	}
	p.scanner.Init(f, text, handleError, 0)
	file := new(ast.File)
	p.next()
	for {
		switch p.tok {
		case token.EOF:
			return file, nil
		case token.FUNC:
			x, err := p.parseFuncDecl()
			if err != nil {
				return nil, err
			}
			file.Funcs = append(file.Funcs, x)
			p.want(token.SEMICOLON)
		default:
			// TODO(kr): don't crash here
			log.Fatalln("parse error", p.pos, p.tok, p.lit)
		}
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
			log.Fatalln(p.fileSet.Position(p.pos), "expected if or {")
		}
	}
	return s
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
	switch t := p.tok; t {
	case token.ADD:
		p.next()
		return &ast.BinaryExpr{X: e, Op: t, Y: p.parseExpr()}
	}
	return e
}

func (p *parser) parseTerm() ast.Expr {
	e := p.parseCall()
	switch t := p.tok; t {
	case token.MUL:
		p.next()
		return &ast.BinaryExpr{X: e, Op: t, Y: p.parseTerm()}
	}
	return e
}

func (p *parser) parseCall() ast.Expr {
	x := p.parseAtom()
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
	}
	// TODO(kr): don't crash here
	log.Fatalln(p.fileSet.Position(p.pos), "error tok =", p.tok, "want ident or literal")
	return nil
}

func (p *parser) parseFuncLit() ast.Expr {
	p.want(token.FUNC)
	p.want(token.LPAREN)
	// TODO(kr): parse params
	params := []*ast.Ident{{"x"}}
	p.want(token.RPAREN)
	body := p.parseBlockStmt()
	return &ast.FuncLit{Params: params, Body: body}
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