/*
Package tree provides functionality for managing a tree data structure tailored for word filtering.

Usage:
	import "github.com/mekavehamichlolay/bad-word-service/tree"

Construct a new Tree:
	tree := tree.NewTree()

Add a word to the Tree with optional restrictions on starting and ending characters:
	err := tree.AddWord(word string, dontStartWith []rune, dontFinishWith []rune)
	- word: The word to be added.
	- dontStartWith: A list of characters the word should not start with.
	- dontFinishWith: A list of characters the word should not end with.
	Returns an error if the word could not be added to the tree.

Check if a text contains any words that are present in the Tree:
	positions := tree.HasWord(text string)
	- text: The text to be checked.
	Returns a slice of pairs of indices representing the start and end positions of found words in the text.

Check if a word is present in the Tree:
	found := tree.Has(word string)
	- word: The word to be checked.
	Returns true if the word is present in the tree, false otherwise.

Special Syntax for Word Searching:
The package supports a special syntax for searching words within a text. This syntax allows for more flexible and advanced word matching.

The special syntax consists of the following elements:

- '^': Denotes the start of a word.
- '$': Denotes the end of a word.
- '[' and ']': Encloses a group of optional characters.
- '?': Indicates that the preceding optional characters are optional and can be present zero or one time.
- ',', '.', '"', "'" " " "_", and '\' are also considered as word characters.

Examples of Special Syntax:

1. "^example": Matches any word that starts with "example".
2. "example$": Matches any word that ends with "example".
3. "[abc]example": Matches any word that contains either 'a', 'b', or 'c', followed by "example".
4. "[abc]?example": Matches any word that contains "example" or contains either 'a', 'b', or 'c', followed by "example".
5. "^[abc]?example$": Matches any word that either starts with 'a', 'b', or 'c', followed by "example" or just starts with "example", and ends there.

Note: The special characters ',', '.', '"', "'" " " "_", and '\' are also considered as word characters.

*/

package tree

import (
	"fmt"

	"github.com/mekavehamichlolay/bad-word-service/utils"
)

type Node struct {
	Children         map[rune]*Node
	IsFullWord       bool
	EndOfWordOnly    bool
	DoesNotStartWith []rune
	DoesNotEndWith   []rune
}

type Tree struct {
	Root *Node
}

func NewTree() TreeInterface {
	return &Tree{Root: &Node{Children: make(map[rune]*Node)}}
}

type TreeInterface interface {
	AddWord(word string, dontStartWith []rune, dontFinishWith []rune) error

	HasWord(text string) [][2]uint

	Has(word string) bool
}

func (t *Tree) AddWord(word string, dontStartWith []rune, dontFinishWith []rune) error {
	if len(word) < 2 {
		return fmt.Errorf("word length must be at least two characters")
	}
	cur := t.Root
	if word[0] == '^' {
		if _, ok := cur.Children[' ']; !ok {
			cur.Children[' '] = &Node{Children: make(map[rune]*Node)}
		}
		cur = cur.Children[' ']
		word = word[1:]
		if len(word) < 2 {
			return fmt.Errorf("word length must be at least two characters")
		}
	}
	endNodes, err := add(word, cur)
	if err != nil {
		return err
	}
	for _, node := range endNodes {
		node.DoesNotStartWith = dontStartWith
		node.DoesNotEndWith = dontFinishWith
	}
	return nil
}
func add(word string, node *Node) ([]*Node, error) {
	var nodes []*Node
	if len(word) == 0 {
		return nil, fmt.Errorf("you passed an empty string")
	}
	char := rune(word[0])
	if utils.IsExpectedAsCharacter(char) {
		if _, ok := node.Children[char]; !ok {
			node.Children[char] = &Node{Children: make(map[rune]*Node)}
		}
		if len(word) == 1 {
			if node.Children[char].IsFullWord {
				return nil, fmt.Errorf("the word %s already exists", word)
			}
			node.Children[char].IsFullWord = true
			nodes = append(nodes, node.Children[char])
			return nodes, nil
		}
		return add(word[1:], node.Children[char])
	}
	if char == '$' && len(word) == 1 {
		node.IsFullWord = true
		node.EndOfWordOnly = true
		nodes = append(nodes, node)
		return nodes, nil
	}
	if char == '[' {
		// multiple options
		if len(word) < 4 {
			return nil, fmt.Errorf("it does not make sense to have less then 2 optional chararcters or a closing bracket and question mark")
		}
		word = word[1:]
		var multiOptionChars []rune
		for word[0] != ']' {
			if len(word) < 2 {
				return nil, fmt.Errorf("you have an open bracket without a closing bracket")
			}
			char = rune(word[0])
			if !utils.IsExpectedAsCharacter(char) {
				return nil, fmt.Errorf("you have a non character in the optional part %s", word)
			}
			multiOptionChars = append(multiOptionChars, char)
			word = word[1:]
		}
		if len(multiOptionChars) < 2 && (len(word) < 2 || word[1] != '?') {
			return nil, fmt.Errorf("you have less than 2 optional characters")
		}
		if len(word) < 2 {
			nodes = appendMultiOptionChars(multiOptionChars, node, true)
			return nodes, nil
		}
		if word[1] == '?' && len(word) == 2 {
			nodes = appendMultiOptionChars(multiOptionChars, node, true)
			node.IsFullWord = true
			nodes = append(nodes, node)
			return nodes, nil
		}
		nodes = appendMultiOptionChars(multiOptionChars, node, false)
		if word[1] == '?' {
			nodes = append(nodes, node)
			word = word[2:]
		} else {
			word = word[1:]
		}
		var returnedNodes []*Node
		for _, n := range nodes {
			returned, err := add(word, n)
			if err != nil {
				return nil, err
			}
			returnedNodes = append(returnedNodes, returned...)
		}
		return returnedNodes, nil
	}
	return nil, fmt.Errorf("you have a non character in your word %s", word)
}
func appendMultiOptionChars(chars []rune, node *Node, isWordEnd bool) []*Node {
	var nodes []*Node
	for _, char := range chars {
		if _, ok := node.Children[char]; !ok {
			node.Children[char] = &Node{Children: make(map[rune]*Node)}
		}
		if isWordEnd {
			node.Children[char].IsFullWord = true
		}
		nodes = append(nodes, node.Children[char])
	}
	return nodes
}

func (t *Tree) HasWord(text string) [][2]uint {
	var result [][2]uint
	for p, char := range text {
		if p == 0 {
			if child, ok := t.Root.Children[' ']; ok {
				if returned := walker(child, text, uint16(p), uint16(p)); returned != 0 {
					result = append(result, [2]uint{uint(p), uint(returned)})
				}
			}
		}
		if !utils.IsExpectedAsCharacter(char) {
			continue
		}
		if returned := walker(t.Root, text, uint16(p), uint16(p)); returned != 0 {
			result = append(result, [2]uint{uint(p), uint(returned)})
		}
	}
	return result
}

func walker(node *Node, text string, startPosition, curPosition uint16) uint16 {
	newNode, ok := node.Children[rune(text[curPosition])]
	if !ok {
		return 0
	}
	if newNode.IsFullWord {
		if newNode.EndOfWordOnly {
			if curPosition == uint16(len(text)-1) || utils.IsEndOfWordSign(rune(text[curPosition+1])) {
				return curPosition
			}
			return 0
		}
		if len(newNode.DoesNotStartWith) > 0 && startPosition > 0 {
			for _, char := range newNode.DoesNotStartWith {
				if rune(text[startPosition-1]) == char {
					return 0
				}
			}
		}
		if len(newNode.DoesNotEndWith) > 0 && curPosition+1 < uint16(len(text)-1) {
			for _, char := range newNode.DoesNotEndWith {
				if rune(text[curPosition+1]) == char {
					return 0
				}
			}
		}
		return curPosition
	}
	if uint16(len(text)-1) == curPosition {
		return 0
	}
	return walker(newNode, text, startPosition, curPosition+1)
}

func (t *Tree) Has(word string) bool {
	return has(t.Root, word) || has(t.Root, " "+word)
}

func has(node *Node, word string) bool {
	char := rune(word[0])
	child, ok := node.Children[char]
	if !ok {
		return false
	}
	if len(word) == 1 {
		return node.Children[char].IsFullWord
	}
	return has(child, word[1:])
}

func Walker(node *Node, text string, startPosition, curPosition uint16) uint16 {
	return walker(node, text, startPosition, curPosition)
}
