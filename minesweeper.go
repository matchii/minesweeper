package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// board
const MINE_TILE int = 9
const EMPTY_TILE int = 0

// mask
const COVERED = 0
const VISIBLE = 1
const FLAGGED = 2

//lookup
const UNKNOWN = 0
const TO_CHECK = 1
const CHECKED = 2

const LETTERS string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Notes here
// TODO Update documentation
// TODO Color output

type grid struct {
	// Stores placement of mines and number of mines on the adjacent tiles
	board [][]int

	// Width of the grid
	width int

	// Height of the grid
	height int

	// Number of mines
	mines int

	// Stores information which tiles user sees and which ones are flagged as (supposedly) mined
	mask [][]int

	// Was mine found?
	mineFound bool

	// Number of mine flags placed
	flags int

	// Slice used to check which tiles should be revealed automatically
	lookup [][]int
}

func main() {
	width, height, mines := GetGameParameters()
	g := BuildGrid(width, height)
	g.PlaceMines(mines)
	g.Print()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%d mines placed. Check which field? (Press Ctrl-C to terminate.)\n", mines)
		input, _ := reader.ReadString('\n')
		column := strings.Index(LETTERS, string(input[0]))
		placeFlag := 0
		if strings.Index(input, "*") >= 0 {
			placeFlag = 1
		}
		line, _ := strconv.Atoi(string(input[1:(utf8.RuneCountInString(input)-1-placeFlag)]))
		if placeFlag == 0 {
			g.UncoverTile(line, column)
		} else {
			g.FlagTileAsMined(line, column)
		}
		g.Print()
		g.CheckLoseCondition()
		g.CheckWinCondition()
	}
}

// GetGameParameters takes parameters of new game from command line.
// Missing parameters are derived from present ones or default values are used.
// Returns width and height of grid and number of mines to place.
func GetGameParameters() (int, int, int) {
	var width, height, mines int
	if len(os.Args) > 1 {
		width, _ = strconv.Atoi(os.Args[1])
	} else {
		width = 10
	}
	if len(os.Args) > 2 {
		height, _ = strconv.Atoi(os.Args[2])
	} else {
		height = width
	}
	if len(os.Args) > 3 {
		mines, _ = strconv.Atoi(os.Args[3])
	} else {
		mines = height*width/8
	}

	return width, height, mines
}

// BuildGrid creates main game object.
// Parameters are: number of columns, number of lines.
// Returns grid structure of given size, no mines placed.
func BuildGrid(width, height int) grid {
	board  := make([][]int, height)
	mask   := make([][]int, height)
	lookup := make([][]int, height)
	for i := 0; i <= height-1; i++ {
		board[i]  = make([]int, width)
		mask[i]   = make([]int, width)
		lookup[i] = make([]int, width)
	}
	g := grid{board, width, height, 0, mask, false, 0, lookup}
	return g
}

func (g *grid) PlaceMines(count int) {
	g.mines = count
	if count > g.width*g.height {
		panic(fmt.Sprintf("Cannot place %d mines on the %dx%d grid.", count, g.width, g.height))
	}
	i := 0
	for {
		l, c := g.GetRandomTile()
		if g.board[l][c] != MINE_TILE {
			g.board[l][c] = MINE_TILE
			i++
		}
		if i >= count {
			break
		}
	}
	for l, line := range g.board {
		for c, code := range line {
			if code != MINE_TILE {
				g.board[l][c] = g.CountMinesAround(l, c)
			}
		}
	}
}

func (g *grid) GetRandomTile() (int, int) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var l, c int = r.Intn(g.height), r.Intn(g.width)
	return l, c
}

func (g *grid) Print() {
	cols := fmt.Sprintf(strings.Join(strings.Split(LETTERS, "")[0:g.width], " "))
	fmt.Printf("%s%s", strings.Repeat(" ", 2+1+1), cols)
	fmt.Println()
	fmt.Println()
	for l, line := range g.board {
		fmt.Printf("%d%s", l, strings.Repeat(" ", 2-(l/10)))
		for c, code := range line {
			if g.mask[l][c] == COVERED {
				fmt.Printf(" .")
				continue
			}
			if g.mask[l][c] == FLAGGED {
				fmt.Printf(" X")
				continue
			}
			if code == MINE_TILE {
				fmt.Printf(" *")
			} else {
				if g.board[l][c] == 0 {
					fmt.Print("  ")
				} else {
					fmt.Printf(" %d", g.board[l][c])
				}
			}
		}
		fmt.Println()
	}
}

func (g *grid) CountMinesAround(l, c int) int {
	var count int
	if l > 0          && c > 0         && g.board[l-1][c-1] == MINE_TILE { count++ }
	if l > 0                           && g.board[l-1][c  ] == MINE_TILE { count++ }
	if l > 0          && c < g.width-1 && g.board[l-1][c+1] == MINE_TILE { count++ }
	if                   c > 0         && g.board[l  ][c-1] == MINE_TILE { count++ }

	if                   c < g.width-1 && g.board[l  ][c+1] == MINE_TILE { count++ }
	if l < g.height-1 && c > 0         && g.board[l+1][c-1] == MINE_TILE { count++ }
	if l < g.height-1                  && g.board[l+1][c  ] == MINE_TILE { count++ }
	if l < g.height-1 && c < g.width-1 && g.board[l+1][c+1] == MINE_TILE { count++ }
	return count
}

func (g *grid) UncoverTile(l, c int) {
	if l >= g.height || c >= g.width {
		panic(fmt.Sprintf("Cannot uncover tile %s%d, out of grid", strings.Split(LETTERS, "")[c], l))
	}
	if g.mask[l][c] == FLAGGED {
		return // This tile is flagged as mined, ignore the move
	}
	if g.board[l][c] == MINE_TILE {
		g.mask[l][c] = VISIBLE
		g.mineFound = true
		return
	}
	g.mask[l][c] = VISIBLE
	if g.board[l][c] == EMPTY_TILE {
		g.UncoverAround(l, c)
		g.UncoverTilesToCheck()
	}
}

func (g *grid) FlagTileAsMined(l, c int) {
	if l >= g.height || c >= g.width {
		panic(fmt.Sprintf("Cannot mark tile %s%d, out of grid", strings.Split(LETTERS, "")[c], l))
	}
	if g.mask[l][c] == FLAGGED {
		g.mask[l][c] = COVERED
		g.lookup[l][c] = UNKNOWN
		g.flags--
	} else {
		g.mask[l][c] = FLAGGED
		g.lookup[l][c] = CHECKED
		g.flags++
	}
}

func (g *grid) UncoverAround(l, c int) {
	g.lookup[l][c] = CHECKED
	for col := l-1; col <= l+1; col++ {
		for row := c-1; row <= c+1; row++ {
			if col < 0 || col > g.height-1 || row < 0 || row > g.width-1 {
				continue // out of grid bounds
			}
			if col == l && row == c {
				continue // central point
			}
			if g.board[col][row] == MINE_TILE {
				g.lookup[col][row] = CHECKED
				continue // don't uncover mines
			}
			if g.mask[col][row] == FLAGGED {
				continue // don't touch flags
			}
			g.mask[col][row] = VISIBLE
			if g.lookup[col][row] == UNKNOWN {
				if g.board[col][row] == EMPTY_TILE {
					g.lookup[col][row] = TO_CHECK
				}
				if g.board[col][row] > EMPTY_TILE {
					g.lookup[col][row] = CHECKED
				}
			}
		}
	}
}

func (g *grid) UncoverTilesToCheck() {
	var goOn bool
	for l, line := range g.lookup {
		for c, tile := range line {
			if tile == TO_CHECK {
				g.UncoverAround(l, c)
				goOn = true
			}
		}
	}
	if goOn {
		g.UncoverTilesToCheck()
	}
}

func (g *grid) CheckLoseCondition() {
	if g.mineFound {
		fmt.Println("Bum! You lose.")
		os.Exit(0)
	}
}

func (g *grid) CheckWinCondition() {
	// We won when all non-mined tiles are uncovered and all mines are flagged
	var uncovered int
	for _, line := range g.mask {
		for _, code := range line {
			if code == VISIBLE {
				uncovered++
			}
		}
	}
	if uncovered == g.width*g.height - g.mines && g.flags == g.mines {
		fmt.Println("You won!")
		os.Exit(0)
	}
}
