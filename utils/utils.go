package utils

import "unicode"

const (
	HYPHEN       = '-'
	SPACE        = ' '
	UNDERSCORE   = '_'
	QUOTE        = '\''
	DOUBLE_QUOTE = '"'
	CAPITAL_A    = 'A'
	CAPITAL_Z    = 'Z'
	SMALL_A      = 'a'
	SMALL_Z      = 'z'
	ALEPH        = 'א'
	TAV          = 'ת'
	ETNAHTA      = '֑'
	QAMATS_QATAN = 'ׇ'
	COMMA        = ','
	DOT          = '.'
)

func IsNiqqud(r rune) bool {
	return r >= ETNAHTA && r <= QAMATS_QATAN
}

func IsHebrewLetter(letter rune) bool {
	return letter >= ALEPH && letter <= TAV
}

func IsCapital(letter rune) bool {
	return letter >= CAPITAL_A && letter <= CAPITAL_Z
}

func IsSmall(letter rune) bool {
	return letter >= SMALL_A && letter <= SMALL_Z
}

func IsUsedAsCharacter(letter rune) bool {
	return letter == QUOTE || letter == HYPHEN || letter == SPACE || letter == DOUBLE_QUOTE || letter == UNDERSCORE
}

func IsEnglishLetter(letter rune) bool {
	return IsCapital(letter) || IsSmall(letter)
}

func IsExpectedAsCharacter(letter rune) bool {
	return IsHebrewLetter(letter) || IsEnglishLetter(letter) || IsUsedAsCharacter(letter) || IsNiqqud(letter)
}

func IsEndOfWordSign(letter rune) bool {
	return unicode.IsSpace(letter) ||
		letter == HYPHEN ||
		letter == QUOTE ||
		letter == DOUBLE_QUOTE ||
		letter == UNDERSCORE ||
		letter == COMMA ||
		letter == DOT ||
		!IsExpectedAsCharacter(letter)
}

func ToUpperCase(letter rune) rune {
	if IsSmall(letter) {
		return letter &^ 32
	}
	return letter
}

func ToLowerCase(letter rune) rune {
	if IsCapital(letter) {
		return letter | 32
	}
	return letter
}
