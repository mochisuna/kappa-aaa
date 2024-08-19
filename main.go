package main

import (
	"math/rand"
	"strings"
	"time"

	"github.com/nsf/termbox-go"
)

const OFFSET_SKY = 0   // 空の開始位置
const OFFSET_KAPPA = 4 // 河童の開始位置
const OFFSET_SEA = 7   // 海の開始位置

// アニメーション構造
var (
	dish         = " ,< = > "
	face         = "(  ' e')"
	body         = " |   #|"
	waterSurface = "~"
	cloud        = "---"
	wave         = "^^^"
)

type Kappa struct {
	l1, l2, l3 string
}

// 河童のアニメーションリスト
var kappaAnim = []Kappa{
	{"", "", ""},                    // 0
	{"", "", dish},                  // 1
	{"", dish, face},                // 2
	{dish, face, body},              // 3
	{dish, face + " < Hello", body}, // 4
}

// こんな感じの動き
var kappaMove = []int{
	0, 0, 0, 1, 2, 3, 3, 3, 3, 3, 4, 4, 4, 4, 3, 3, 2, 1,
}

// 乱数生成
var random = rand.New(rand.NewSource(time.Now().UnixNano()))

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	width, _ := termbox.Size()
	offset := 0
	kappaMoveIndex := 0
	offsetSky := make([][]int, 3)
	for i := 0; i < len(offsetSky); i++ {
		offsetSky[i] = make([]int, 5)
	}
	offsetSea := make([][]int, 5)
	for i := 0; i < len(offsetSea); i++ {
		offsetSea[i] = make([]int, 5)
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	eventQueue := make(chan termbox.Event)
	go func() {
		for {
			eventQueue <- termbox.PollEvent()
		}
	}()

	for {
		select {
		case ev := <-eventQueue:
			if ev.Type == termbox.EventKey {
				if ev.Key == termbox.KeyEnter || ev.Ch == 'q' {
					return
				}
			}
		case <-ticker.C:
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			if offset%16 == 0 {
				makeRandomOffset(&offsetSky)
			}
			if offset%9 == 0 {
				makeRandomOffset(&offsetSea)
			}
			drawKappa(kappaMoveIndex, offset)
			drawSky(offsetSky)
			drawSea(offsetSea)
			termbox.Flush()

			// 画面端にいったら初期位置にワープ
			offset++
			if offset >= width {
				offset = 0
			}
			kappaMoveIndex = (kappaMoveIndex + 1) % len(kappaMove)
		}
	}
}

// 雲と波をランダム生成
func makeRandomOffset(offset *[][]int) {
	width, _ := termbox.Size()
	for i := 0; i < len((*offset)); i++ {
		for j := 0; j < len((*offset)[i]); j++ {
			(*offset)[i][j] = random.Intn(width)
		}
	}
}

func drawSky(offsetSky [][]int) {
	for i := 0; i < len(offsetSky); i++ {
		for j := 0; j < len(offsetSky[0]); j++ {
			draw(offsetSky[i][j], OFFSET_SKY+i, cloud)
		}
	}
}
func drawKappa(index, offset int) {
	draw(offset, OFFSET_KAPPA, kappaAnim[kappaMove[index]].l1)
	draw(offset, OFFSET_KAPPA+1, kappaAnim[kappaMove[index]].l2)
	draw(offset, OFFSET_KAPPA+2, kappaAnim[kappaMove[index]].l3)

}

func drawSea(offsetSea [][]int) {
	width, _ := termbox.Size()
	// 画面幅いっぱいに波線を描画
	surface := strings.Repeat(waterSurface, width)
	draw(0, OFFSET_SEA, surface)
	for i := 0; i < len(offsetSea); i++ {
		for j := 0; j < len(offsetSea[0]); j++ {
			// 水面の分だけ一個ずらす
			draw(offsetSea[i][j], OFFSET_SEA+1+i, wave)
		}
	}
}

// 規定のx, yに文字を表示
func draw(x, y int, s string) {
	for i, r := range s {
		termbox.SetCell(x+i, y, r, termbox.ColorWhite, termbox.ColorDefault)
	}
}
