package main

import (
	"flag"
	"math/rand"
	"strings"
	"time"

	"github.com/nsf/termbox-go"
)

const OFFSET_SKY = 0                               // 空の開始位置
const HEIGHT_SKY = 3                               // 空の高さ
const OFFSET_KAPPA = HEIGHT_SKY + 2                // 河童の開始位置（空の高さに余白を入れた次）
const HEIGHT_KAPPA = 4                             // 河童の高さ
const OFFSET_SEA = OFFSET_KAPPA + HEIGHT_KAPPA - 2 // 海の開始位置（ちょっとだけ河童が海にめり込む）
const HEIGHT_SEA = 4                               // 海の高さ
const COUNT_CLOUD = 5                              // 雲の個数
const COUNT_WAVE = 8                               // 波の個数

// アニメーション構造
var (
	dish         = " ,< = > "
	face         = "(  ' e')"
	body         = " |   #|"
	foot         = "~~~~~~~~"
	waterSurface = "~"
	cloud        = "---"
	wave         = "^^^^^"
)

// 乱数生成器
var random = rand.New(rand.NewSource(time.Now().UnixNano()))

// 河童はレイヤーが徐々に切り替わるので先にアニメーション定義しておく
type kappaFrame struct {
	l1, l2, l3, l4 string
}

// 河童のアニメーション全体を取り扱う構造体
type KappaAnim struct {
	frames  []kappaFrame
	actions []int
	index   int
}

type SkySea struct {
	offset [][]int
}

func NewKappaAnim(msg string) KappaAnim {
	// 河童のアニメーションリスト
	frames := []kappaFrame{
		{"", "", "", ""},                       // 0
		{"", "", dish, foot},                   // 1
		{"", dish, face, foot},                 // 2
		{dish, face, body, foot},               // 3
		{dish, face + " < " + msg, body, foot}, // 4
	}

	// こんな感じの動き
	actions := []int{
		0, 0, 0, 1, 2, 3, 3, 3, 3, 3, 4, 4, 4, 4, 3, 3, 2, 1,
	}
	return KappaAnim{
		frames:  frames,
		actions: actions,
		index:   0,
	}
}

// 空と海は構造が同じなので面倒だから一緒くたに定義
func NewSkySea(h, c int) SkySea {
	offset := make([][]int, h)
	for i := 0; i < len(offset); i++ {
		offset[i] = make([]int, c)
	}
	return SkySea{offset}
}

// 雲と波をランダム生成
func (ss *SkySea) makeRandomOffset() {
	width, _ := termbox.Size()
	for i := 0; i < len(ss.offset); i++ {
		for j := 0; j < len(ss.offset[i]); j++ {
			ss.offset[i][j] = random.Intn(width)
		}
	}
}

func (k *KappaAnim) getFrame() kappaFrame {
	return k.frames[k.actions[k.index]]
}

func (k *KappaAnim) drawKappa(offset int) {
	k.index = (k.index + 1) % len(k.actions)
	frame := k.getFrame()
	draw(offset, OFFSET_KAPPA, frame.l1)
	draw(offset, OFFSET_KAPPA+1, frame.l2)
	draw(offset, OFFSET_KAPPA+2, frame.l3)
	draw(offset, OFFSET_KAPPA+3, frame.l4)
}

func (ss *SkySea) drawSky() {
	for i := 0; i < len(ss.offset); i++ {
		for j := 0; j < len(ss.offset[0]); j++ {
			draw(ss.offset[i][j], OFFSET_SKY+i, cloud)
		}
	}
}
func (ss *SkySea) drawSea() {
	width, _ := termbox.Size()
	// 画面幅いっぱいに波線を描画
	surface := strings.Repeat(waterSurface, width)
	draw(0, OFFSET_SEA, surface)
	for i := 0; i < len(ss.offset); i++ {
		for j := 0; j < len(ss.offset[0]); j++ {
			// 水面と河童の分だけ2個ずらす
			draw(ss.offset[i][j], OFFSET_SEA+2+i, wave)
		}
	}
}

// 規定のx, yに文字を表示
func draw(x, y int, s string) {
	for i, r := range s {
		termbox.SetCell(x+i, y, r, termbox.ColorWhite, termbox.ColorDefault)
	}
}

func main() {
	msg := flag.String("m", "Hello!", "Define the message that Kappa speaks.")
	flag.Parse()

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	width, _ := termbox.Size()
	offset := 0
	kappa := NewKappaAnim(*msg)
	sky := NewSkySea(HEIGHT_SKY, COUNT_CLOUD)
	sea := NewSkySea(HEIGHT_SEA, COUNT_WAVE)

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
		// qかspaceが入力されたらやめる
		case ev := <-eventQueue:
			if ev.Type == termbox.EventKey {
				if ev.Key == termbox.KeyEnter || ev.Ch == 'q' {
					return
				}
			}
		case <-ticker.C:
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			if offset%17 == 0 {
				sky.makeRandomOffset()
			}
			if offset%8 == 0 {
				sea.makeRandomOffset()
			}
			sky.drawSky()
			sea.drawSea()
			kappa.drawKappa(offset)
			termbox.Flush()

			// 画面端にいったら初期位置にワープ
			// slコマンドと同じにするならここでexit
			offset++
			if offset >= width {
				offset = 0
			}
		}
	}
}
