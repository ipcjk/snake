// snake in golang / linux macos terminal Jörg Kost, jk@ip-clear.de

package main

import (
	"bufio"
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"math/rand"
	"os"
	"time"
)

const (
	empty = iota
	fence
	snake
	food
)

var gameField = [25][80]int{}

func main() {
	const maxSnakeLength = 256
	var score int
	var lastAction rune

	gameOver := false
	inputKeys := make(chan rune)
	rand.Seed(time.Now().UnixNano())
	snakeElements := [maxSnakeLength + 1][1][2]int{}

	snakeHeadPosX, snakeHeadPosY := 12, 40
	snakeElements[0][0][0], snakeElements[0][0][1] = snakeHeadPosX, snakeHeadPosY

	/* parse flags */
	gameSpeed := flag.Duration("s", 30000000, "speed")
	playerName := flag.String("p", "Nibbles", "name of the player")
	randomElements := flag.Int("r", 0, "numbers of random elements to be placed on gamefield")
	snakeLength := flag.Int("l", 0, "starting length of nibbles")
	flag.Parse()

	for i := 0; i < *randomElements; i++ {
		placeThing(randomXPos(), randomYPos(), fence)
		placeThing(randomXPos(), randomYPos(), food)
	}

	/* make terminal raw so we can read input chars from */
	state, err := terminal.MakeRaw(0)
	if err != nil {
		log.Fatalln("setting stdin to raw:", err)
	}
	defer func() {
		fmt.Print("\033[?25h")
		if err := terminal.Restore(0, state); err != nil {
			log.Fatal(err)
		}
	}()

	/* init gamefield */
	for x := 0; x < 25; x++ {
		for y := 0; y < 80; y++ {
			if x == 1 || x == 24 {
				gameField[x][y] = fence
			}
			if y == 1 || y == 79 {
				gameField[x][y] = fence
			}
		}
	}

	/* random food placement */
	placeThing(randomXPos(), randomYPos(), food)

	print("\033[H\033[2J")
	print("\033[?25l")
	printPlayerName(*playerName)

	// Draw init field
	for x := 0; x < 25; x++ {
		for y := 0; y < 80; y++ {
			fmt.Printf("\033[%d;%dH", x, y)
			switch gameField[x][y] {
			case fence:
				fmt.Print("#")
			case food:
				fmt.Printf("\x1b[32m%s\x1b[0m", "*")
			}
		}
	}

	/* read runes in a go thread and pushes it into a channel */
	go func() {
		in := bufio.NewReader(os.Stdin)
		for {
			r, _, _ := in.ReadRune()
			inputKeys <- r
			if r == 'q' {
				break
			}
		}
	}()

	/* Game loop */
	lastAction = 'x'
	var beforeElement int
	for !gameOver {
		/* time to react*/
		time.Sleep(*gameSpeed)

		/* rotate old snake elements to the right side */
		for i := *snakeLength + 1; i >= 0; i-- {
			snakeElements[i+1][0][0] = snakeElements[i][0][0]
			snakeElements[i+1][0][1] = snakeElements[i][0][1]
		}
		/* Save new head coordinates */
		snakeElements[0][0][0] = snakeHeadPosX
		snakeElements[0][0][1] = snakeHeadPosY

		/* print new snake head coordinates */
		printCoord(snakeHeadPosX, snakeHeadPosY)

		/* Save element from the gamefield that was on this position, e.g. food, ...
		only do if we are not in the startup phase */
		if lastAction != 'x' {
			beforeElement = gameField[snakeHeadPosX][snakeHeadPosY]
			/* overwrite element with snake head */
		}

		placeThing(snakeHeadPosX, snakeHeadPosY, snake)

		/* now clear the "old" tail by moving the cursor inside multidimensional array + 1, if there has been
		a snake on this gamefield coords, we need to clear it with an empty graphic element
		*/
		if lastAction != 'x' && gameField[snakeElements[*snakeLength+1][0][0]][snakeElements[*snakeLength+1][0][1]] == snake {
			placeThing(snakeElements[*snakeLength+1][0][0], snakeElements[*snakeLength+1][0][1], empty)
		}

		/* check input keys channel */
		select {
		case r, ok := <-inputKeys:
			if ok {
				switch r {
				case 'w':
					if lastAction != 's' {
						lastAction = r
					}
				case 'a':
					if lastAction != 'd' {
						lastAction = r
					}
				case 's':
					if lastAction != 'w' {
						lastAction = r
					}
				case 'd':
					if lastAction != 'a' {
						lastAction = r
					}
				case 'q':
					gameOver = true
					break
				}
			}
		default:
			break
		}

		switch beforeElement {
		case fence:
			gameOver = true
			break
		case food:
			if *snakeLength+1 < maxSnakeLength {
				*snakeLength++
			}
			score += 10
			printScore(*snakeLength, score)
			placeThing(randomXPos(), randomYPos(), food)
		case snake:
			if lastAction != 'x' {
				gameOver = true
			}
			/* if the game starts, we got an auto "snake" field but an lastAction == x
			so we dont quit then
			*/
		}

		/* calculate next position for the snake head */
		switch lastAction {
		case 'w':
			snakeHeadPosX = snakeElements[0][0][0] - 1
		case 's':
			snakeHeadPosX = snakeElements[0][0][0] + 1
		case 'd':
			snakeHeadPosY = snakeElements[0][0][1] + 1
		case 'a':
			snakeHeadPosY = snakeElements[0][0][1] - 1
		}
	}
}

func randomXPos() (randomX int) {
	randomX = rand.Intn(24)
	if randomX == 1 {
		randomX++
	}
	return
}

func randomYPos() (randomY int) {
	randomY = rand.Intn(78)
	if randomY == 0 {
		randomY++
	}
	return
}

func placeThing(x, y, thing int) {
	if thing != empty && gameField[x][y] == snake {
		return
	}
	gameField[x][y] = thing
	fmt.Printf("\033[%d;%dH", x, y)
	switch thing {
	case fence:
		fmt.Print("#")
	case food:
		fmt.Printf("\x1b[32m%s\x1b[0m", "*")
	case snake:
		fmt.Printf("\x1b[31m%s\x1b[0m", "S")
	case empty:
		fmt.Print(" ")
	}
}
func printScore(length int, score int) {
	fmt.Printf("\033[%d;%dH L: %d  Score: %d", 25, 0, length, score)
}
func printPlayerName(msg string) {
	fmt.Printf("\033[%d;%dH Player: ", 25, 22)
	fmt.Printf("%s", msg)
}
func printCoord(x, y int) {
	fmt.Printf("\033[%d;%dH Coord: ", 25, 50)
	fmt.Printf("  %d  ,   %d   ", x, y)
}
