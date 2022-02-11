package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/nsf/termbox-go"
	"golang.org/x/term"
)

// CONSTANTS ========================================================
const UP = "w"
const DOWN = "s"
const LEFT = "a"
const RIGHT = "d"
const PAUSE = "p"
const ESC = "q"
const F_POINTS = 10
const S_POINTS = 100

const BORDER = "*" // "*"
const HEAD = "o"   // "\033[31;1;96mo\033[0m"
const BODY = "x"   // "\033[31;1;96mx\033[0m"
const F = "F"      // "\033[31;1;91mF\033[0m"
const S = "S"      // "\033[31;1;91mS\033[0m"

// STRUCTURES =======================================================
type board struct {
	xy [][]string
}

type node struct {
	x, y int
	next *node
}
type snake struct {
	hx, hy int // snake direction
	first  *node
}

type fruit struct {
	x, y  int
	value string
}

// GLOBAL VARIABLES =================================================
var OS = "" // Operating System
var BL = 0  // Board Length
var BH = 0  // Board Height

var game_over = false
var tot_points = 0

var b *board = new(board)
var s *snake = new(snake)
var f *fruit = new(fruit)

var input_channel = make(chan string, 5)

// FUNCTIONS ========================================================
func init_board() {
	b.xy = make([][]string, BL)
	for i := range b.xy {
		b.xy[i] = make([]string, BH)
	}

	for i := 0; i < BL; i++ {
		for j := 0; j < BH; j++ {
			if i == 0 || i == BL-1 || j == 0 || j == BH-1 {
				b.xy[i][j] = "*"
			} else {
				b.xy[i][j] = " "
			}
		}
	}
}

// init the snake with length of 4, in the center of the screen
func init_snake() {
	s.hx = -1
	s.hy = 0
	n4 := node{x: BL/2 + 3, y: BH / 2, next: nil}
	n3 := node{x: BL/2 + 2, y: BH / 2, next: &n4}
	n2 := node{x: BL/2 + 1, y: BH / 2, next: &n3}
	n1 := node{x: BL / 2, y: BH / 2, next: &n2}
	s.first = &n1
}

// spawn fruit in a random position inside the board
func spawn_fruit() {
	f.x = rand.Intn(BL-2) + 1
	f.y = rand.Intn(BH-2) + 1
	if rand.Intn(100) < 10 {
		f.value = S
	} else {
		f.value = F
	}
	b.xy[f.x][f.y] = f.value
}

// add a snake node on the head
func add_node(x, y int) {
	n := node{x: x, y: y, next: s.first}
	s.first = &n
}

func update_snake_position() {
	// checks for collision with snake
	if b.xy[s.first.x+s.hx][s.first.y+s.hy] == BODY {
		game_over = true
		return
	}
	add_node(s.first.x+s.hx, s.first.y+s.hy)

	// checks for collision with borders
	if s.first.x == 0 || s.first.x == BL-1 || s.first.y == 0 || s.first.y == BH-1 {
		game_over = true
		return
	}

	prev_node := s.first
	node := s.first.next
	for {
		if node.next != nil {
			prev_node = node
			node = node.next
		} else {
			b.xy[node.x][node.y] = " "
			prev_node.next = nil
			break
		}
	}
}

// updates the snake inside the board
func update_board() {
	var node *node

	node = s.first.next
	b.xy[s.first.x][s.first.y] = HEAD

	for {
		b.xy[node.x][node.y] = BODY
		if node.next != nil {

			node = node.next
		} else {
			break
		}
	}
}

// check if the snake has collected a fruit
func collect_fruit() {
	if s.first.x == f.x && s.first.y == f.y {
		if f.value == F {
			tot_points += F_POINTS
		} else if f.value == S {
			tot_points += S_POINTS
		} else {
			// Error
			return
		}
		add_node(f.x, f.y)
		spawn_fruit()
	}
}

func draw() {
	fmt.Printf("\033[1;0H")
	for y := 0; y < BH; y++ {
		for x := 0; x < BL; x++ {
			fmt.Printf("%s", b.xy[x][y])
		}
		fmt.Printf("\n")
	}
}

func print_game_over() {
	fmt.Print("\033[H\033[2J")
	for j := 0; j < BH; j++ {
		for i := 0; i < BL; i++ {
			switch {
			case i == 0 || i == BL-1 || j == 0 || j == BH-1:
				fmt.Printf("*")
			case j == BH/2 && i == BL/2-5:
				fmt.Printf("GAME OVER")
				i += 8
			case j == BH/2+2 && i == BL/2-(8+(len(strconv.Itoa(tot_points))/2)):
				fmt.Printf("Total Points: %d", tot_points)
				i += 13 + len(strconv.Itoa(tot_points))
			default:
				fmt.Printf(" ")
			}
		}
		fmt.Printf("\n")
	}
}

func show_points() {
	fmt.Printf("Points: %d", tot_points)
}

// GOROUTINES =======================================================
func game() {

	for !game_over {
		update_snake_position()
		update_board()
		collect_fruit()
		draw()
		show_points()
		// check if there are inputs
		select {
		case x := <-input_channel:
			switch x {
			case UP:
				if s.hy != 1 {
					s.hx = 0
					s.hy = -1
				}
			case DOWN:
				if s.hy != -1 {
					s.hx = 0
					s.hy = +1
				}
			case LEFT:
				if s.hx != 1 {
					s.hx = -1
					s.hy = 0
				}
			case RIGHT:
				if s.hx != -1 {
					s.hx = +1
					s.hy = 0
				}
			case PAUSE:
				b.xy[BL/2-4][BH/2-1] = " "
				b.xy[BL/2-3][BH/2-1] = " "
				b.xy[BL/2-2][BH/2-1] = " "
				b.xy[BL/2-1][BH/2-1] = " "
				b.xy[BL/2][BH/2-1] = " "
				b.xy[BL/2+1][BH/2-1] = " "
				b.xy[BL/2+2][BH/2-1] = " "
				b.xy[BL/2-4][BH/2] = " "
				b.xy[BL/2-3][BH/2] = "P"
				b.xy[BL/2-2][BH/2] = "A"
				b.xy[BL/2-1][BH/2] = "U"
				b.xy[BL/2][BH/2] = "S"
				b.xy[BL/2+1][BH/2] = "E"
				b.xy[BL/2+2][BH/2] = " "
				b.xy[BL/2-4][BH/2+1] = " "
				b.xy[BL/2-3][BH/2+1] = " "
				b.xy[BL/2-2][BH/2+1] = " "
				b.xy[BL/2-1][BH/2+1] = " "
				b.xy[BL/2][BH/2+1] = " "
				b.xy[BL/2+1][BH/2+1] = " "
				b.xy[BL/2+2][BH/2+1] = " "
				draw()
				x = <-input_channel
				if x == ESC {
					return
				}
				b.xy[BL/2-3][BH/2] = " "
				b.xy[BL/2-2][BH/2] = " "
				b.xy[BL/2-1][BH/2] = " "
				b.xy[BL/2][BH/2] = " "
				b.xy[BL/2+1][BH/2] = " "
			case ESC:
				return
			default:
				fmt.Printf("[INPUT] Input %s not valid. Press 'q' to quit\n", x)
			}
		default:
			continue
		}
	}
}

func input_sampler() {
	// switch stdin into 'raw' mode
	switch {
	case OS == "windows":
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Println(err)
			return
		}
		defer term.Restore(int(os.Stdin.Fd()), oldState)
	case OS == "darwin" || OS == "linux":
		exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run() // disable input buffering
		exec.Command("stty", "-F", "/dev/tty", "-echo").Run()              // do not display entered characters on the screen
		defer exec.Command("stty", "-F", "/dev/tty", "echo").Run()
	default:
		fmt.Printf("%s.\n", OS)
		return
	}
	ch := make([]byte, 1)
	var err error
	for {
		// read byte
		_, err = os.Stdin.Read(ch)
		if err != nil {
			fmt.Println(err)
			return
		}

		// send on channel
		input_channel <- string(ch[0])

	}
}

// MAIN =============================================================
func main() {
	fmt.Printf("")
	// Init rand seed
	rand.Seed(time.Now().UnixNano())

	// Init termbox
	if err := termbox.Init(); err != nil {
		panic(err)
	}

	// Check OS
	OS = runtime.GOOS
	// switch stdin into 'raw' mode
	switch {
	case OS == "windows":
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Println(err)
			return
		}
		defer term.Restore(int(os.Stdin.Fd()), oldState)
	case OS == "darwin" || OS == "linux":
		exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run() // disable input buffering
		exec.Command("stty", "-F", "/dev/tty", "-echo").Run()              // do not display entered characters on the screen
		defer exec.Command("stty", "-F", "/dev/tty", "echo").Run()
	default:
		fmt.Printf("%s.\n", OS)
		return
	}

	fmt.Printf("\033[H\033[2J\n---- CONTROLS ----\nw = up\ns = down\na = left\nd = right\n\np = pause\nq = quit\n\n\nChoose the difficulty by resizing the window.\nSmaller window leads to smaller board;\nfaster snake, bigger window leads to bigger board and slower snake.\n\n\n\nPress any key to start ...")
	ch := make([]byte, 1)
	_, err := os.Stdin.Read(ch)
	if err != nil {
		fmt.Println(err)
		return
	}
	if string(ch[0]) == "q" {
		return
	}

	w, h := termbox.Size()
	termbox.Close()
	fmt.Println(w, h) // test
	BL = w
	BH = h - 1

	init_board()
	init_snake()
	spawn_fruit()

	go input_sampler()
	game()

	print_game_over()
}
