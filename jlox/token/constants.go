package token

type Type int

const (
	_ Type = iota

	// Single-character tokens.

	LeftParen
	RightParen
	LeftBrace
	RightBrace
	Comma
	Dot
	Minus
	Plus
	Semicolon
	Slash
	Star

	// One/two character tokens

	Bang
	BangEqual
	Equal
	EqualEqual
	Greater
	GreaterEqual
	Less
	LessEqual

	// Literals

	Identifier
	String
	Number

	// Keywords

	And
	Class
	Else
	False
	For
	Fun
	If
	Nil
	Or
	Print
	Return
	Super
	This
	True
	Var
	While

	EOF
)

var tokenNames map[Type]string

func (t Type) String() string {
	name, ok := tokenNames[t]
	if !ok {
		return "TOK_UNKNOWN"
	}

	return name
}

func init() {
	tokenNames = map[Type]string{
		LeftParen:    "TOK_LEFT_PAREN",
		RightParen:   "TOK_RIGHT_PAREN",
		LeftBrace:    "TOK_LEFT_BRACE",
		RightBrace:   "TOK_RIGHT_BRACE",
		Comma:        "TOK_COMMA",
		Dot:          "TOK_DOT",
		Minus:        "TOK_MINUS",
		Plus:         "TOK_PLUS",
		Semicolon:    "TOK_SEMICOLON",
		Slash:        "TOK_SLASH",
		Star:         "TOK_STAR",
		Bang:         "TOK_BANG",
		BangEqual:    "TOK_BANG_EQUAL",
		Equal:        "TOK_EQUAL",
		EqualEqual:   "TOK_EQUAL_EQUAL",
		Greater:      "TOK_GREATER",
		GreaterEqual: "TOK_GREATER_EQUAL",
		Less:         "TOK_LESS",
		LessEqual:    "TOK_LESS_EQUAL",
		Identifier:   "TOK_IDENTIFIER",
		String:       "TOK_STRING",
		Number:       "TOK_NUMBER",
		And:          "TOK_AND",
		Class:        "TOK_CLASS",
		Else:         "TOK_ELSE",
		False:        "TOK_FALSE",
		For:          "TOK_FOR",
		Fun:          "TOK_FUN",
		If:           "TOK_IF",
		Nil:          "TOK_NIL",
		Or:           "TOK_OR",
		Print:        "TOK_PRINT",
		Return:       "TOK_RETURN",
		Super:        "TOK_SUPER",
		This:         "TOK_THIS",
		True:         "TOK_TRUE",
		Var:          "TOK_VAR",
		While:        "TOK_WHILE",
		EOF:          "TOK_EOF",
	}
}
