package maptree

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"sync"

	"github.com/mekavehamichlolay/bad-word-service/utils"
)

type Node struct {
	DontStartWith, DontEndWith     []rune
	EndOfWordOnly, StartOfWordOnly bool
}
type Tree struct {
	Children map[string]*Node
	Sizes    []int
	mutex    sync.RWMutex
}

type TreeInterface interface {
	AddWord(word []rune, dontStartWith []rune, dontFinishWith []rune) error
	HasWord(text string) [][2]uint
	Has(word string) bool
	Reset(ctx context.Context, conn *sql.Conn) error
	set(res *sql.Rows) error
}

func NewTree() *Tree {
	return &Tree{Children: make(map[string]*Node)}
}

func (t *Tree) AddWord(word []rune, dontStartWith []rune, dontFinishWith []rune) error {
	if len(word) < 2 {
		return fmt.Errorf("word length must be at least two characters")
	}
	startOfWordOnly := false
	if word[0] == '^' {
		startOfWordOnly = true
		word = word[1:]
		if len(word) < 2 {
			return fmt.Errorf("word length must be at least two characters")
		}
	}
	endNodes, err := add([][]rune{}, word)
	if err != nil {
		return err
	}
	t.mutex.Lock()
	defer t.mutex.Unlock()
	for _, node := range endNodes {
		endOfWordOnly := node[len(node)-1] == ' '
		if endOfWordOnly {
			node = node[:len(node)-1]
		}
		if _, ok := t.Children[string(node)]; ok {
			return fmt.Errorf("word already exists")
		}
		t.Children[string(node)] = &Node{
			DontStartWith:   dontStartWith,
			DontEndWith:     dontFinishWith,
			EndOfWordOnly:   endOfWordOnly,
			StartOfWordOnly: startOfWordOnly,
		}
		t.SetSize(len(node))
	}
	return nil
}

func add(words [][]rune, word []rune) ([][]rune, error) {
	if len(word) == 0 {
		return nil, fmt.Errorf("you passed an empty string")
	}
	char := utils.ToLowerCase(word[0])
	if utils.IsExpectedAsCharacter(char) {
		for i := 0; i < len(words); i++ {
			words[i] = append(words[i], char)
		}
		if len(words) == 0 {
			words = append(words, []rune{char})
		}
		if len(word) == 1 {
			return words, nil
		}
		return add(words, word[1:])
	}
	if char == '$' && len([]rune(word)) == 1 {
		for i := 0; i < len(words); i++ {
			words[i] = append(words[i], ' ')
		}
		return words, nil
	}
	if char == '[' {
		// multiple options
		if len([]rune(word)) < 4 {
			return nil, fmt.Errorf("it does not make sense to have less then 2 optional chararcters or a closing bracket and question mark")
		}
		word = word[1:]
		var multiOptionChars []rune
		for word[0] != ']' {
			if len(word) < 2 {
				return nil, fmt.Errorf("you have an open bracket without a closing bracket")
			}
			char = utils.ToLowerCase(word[0])
			if !utils.IsExpectedAsCharacter(char) {
				return nil, fmt.Errorf("you have a non character in the optional part %s", string(word))
			}
			multiOptionChars = append(multiOptionChars, char)
			word = word[1:]
		}
		if len(multiOptionChars) < 2 && (len(word) < 2 || word[1] != '?') {
			return nil, fmt.Errorf("you have less than 2 optional characters")
		}
		if len(word) < 2 {
			var newWords [][]rune
			for i := 0; i < len(words); i++ {
				for in := 0; in < len(multiOptionChars); in++ {
					newWords = append(newWords, []rune(string(words[i])+string(multiOptionChars[in])))
				}
			}
			return newWords, nil
		}
		if word[1] == '?' && len(word) == 2 {
			var newWords [][]rune
			for i := 0; i < len(words); i++ {
				for in := 0; in < len(multiOptionChars); in++ {
					newWords = append(newWords, []rune(string(words[i])+string(multiOptionChars[in])))
				}
			}
			words = append(words, newWords...)
			return words, nil
		}
		var newWords [][]rune
		for i := 0; i < len(words); i++ {
			for in := 0; in < len(multiOptionChars); in++ {
				newWords = append(newWords, []rune(string(words[i])+string(multiOptionChars[in])))
			}
		}
		if word[1] == '?' {
			words = append(words, newWords...)
			word = word[2:]
		} else {
			words = newWords
			word = word[1:]
		}
		return add(words, word)
	}
	return nil, fmt.Errorf("you have a non character in the optional part %s", string(word))
}

func (t *Tree) SetSize(size int) {
	// dont lock. if where here its already locked
	for i := 0; i < len(t.Sizes); i++ {
		if t.Sizes[i] == size {
			return
		}
	}
	t.Sizes = append(t.Sizes, size)
}
func (t *Tree) HasWord(text string) [][2]uint {
	var result [][2]uint
	runeText := []rune(text)
	t.mutex.RLock()
	for _, length := range t.Sizes {
	wordWalker:
		for i := 0; i < len(runeText)-length; i++ {
			word := string(runeText[i : i+length])
			if node, ok := t.Children[word]; ok {
				if node.StartOfWordOnly {
					if i != 0 && !isStartOrEndOfWord(runeText[i-1]) {
						continue wordWalker
					}
				}
				if node.EndOfWordOnly {
					if i+length+1 < len(runeText) && !isStartOrEndOfWord(runeText[i+length]) {
						continue wordWalker
					}
				}
				if node.DontStartWith != nil && i != 0 {
					for _, c := range node.DontStartWith {
						if runeText[i-1] == c {
							continue wordWalker
						}
					}
				}
				if node.DontEndWith != nil && i+length+1 != len(runeText) {
					for _, c := range node.DontEndWith {
						if runeText[i+length] == c {
							continue wordWalker
						}
					}
				}
				result = append(result, [2]uint{uint(i), uint(i + length)})
			}
		}
	}
	t.mutex.RUnlock()
	slices.SortStableFunc(result, func(a, b [2]uint) int {
		return int(a[0]) - int(b[0])
	})
	return result
}
func isStartOrEndOfWord(c rune) bool {
	switch c {
	case ' ', '\n', '\t', '\r', '|', '!', '?', '.', ',', ';', ':', '(', ')', '[', ']', '{', '}', '<', '>', '/', '\\', '%', '@', '&', '*', '^', '+', '-', '_', '=', '~', '`':
		return true
	}
	return false
}

func (t *Tree) Has(word string) bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	_, ok := t.Children[word]
	return ok
}

func (tree *Tree) Reset(ctx context.Context, conn *sql.Conn) error {
	defer conn.Close()
	rows, err := conn.QueryContext(ctx, "SELECT bw_word, bw_dont_start_with, bw_dont_end_with FROM mw_bad_words")
	if err != nil {
		return err
	}
	defer rows.Close()
	if err := tree.set(rows); err != nil {
		return err
	}
	return nil
}

func (t *Tree) Set(words [][3][]rune) error {
	var errores []error = make([]error, 0)
	for _, word := range words {
		if err := t.AddWord(word[0], word[1], word[2]); err != nil {
			errores = append(errores, err)
		}
	}
	if len(errores) > 0 {
		return fmt.Errorf("errors: %v", errores)
	}
	return nil
}

func (t *Tree) set(res *sql.Rows) error {
	t.mutex.Lock()
	t.Children = nil
	t.Sizes = nil
	t.Children = make(map[string]*Node)
	t.Sizes = make([]int, 0)
	t.mutex.Unlock()
	for res.Next() {
		var bw badWord
		if err := res.Scan(&bw.word, &bw.dontStartWith, &bw.dontEndWith); err != nil {
			return err
		}
		if err := t.AddWord([]rune(bw.word), []rune(bw.dontStartWith), []rune(bw.dontEndWith)); err != nil {
			return err
		}
	}
	if err := res.Close(); err != nil {
		return err
	}
	if err := res.Err(); err != nil {
		return err
	}
	return nil
}
