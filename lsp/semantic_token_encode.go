package lsp

type SemanticToken struct {
	line           int
	startChar      int
	length         int
	tokenType      int
	tokenModifiers int
}

type SemanticTokenBuilder struct {
	tokens []SemanticToken
}

func (b *SemanticTokenBuilder) AddToken(line, startChar, length, tokenType, tokenModifiers int) {
	b.tokens = append(b.tokens, SemanticToken{
		line:           line,
		startChar:      startChar,
		length:         length,
		tokenType:      tokenType,
		tokenModifiers: tokenModifiers,
	})
}

func (b *SemanticTokenBuilder) Build() []uint {
	data := []uint{}
	for i := range b.tokens {
		encoded := b.encodeTokenAt(i)
		data = append(data, encoded...)
	}
	return data
}

func (b *SemanticTokenBuilder) encodeTokenAt(index int) []uint {
	token := b.tokens[index]
	encoded := make([]uint, 5)

	prevLine := 0
	prevStartChar := 0
	if index > 0 {
		prevToken := b.tokens[index-1]
		prevLine = prevToken.line
		if token.line == prevLine {
			prevStartChar = prevToken.startChar
		}
	}

	lineDelta := token.line - prevLine
	startCharDelta := token.startChar - prevStartChar
	encoded[0] = uint(lineDelta)
	encoded[1] = uint(startCharDelta)
	encoded[2] = uint(token.length)
	encoded[3] = uint(token.tokenType)
	encoded[4] = uint(token.tokenModifiers)

	return encoded
}
