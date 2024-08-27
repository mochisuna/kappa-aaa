package main

import (
	"flag"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/nsf/termbox-go"
)

const (
	offsetSky     = iota                      // 空の開始位置
	heightSky     = 3                         // 空の高さ
	offsetKappa   = heightSky + 2             // 河童の開始位置（空の高さに余白を入れた次）
	heightKappa   = 4                         // 河童の高さ
	offsetSea     = offsetKappa + heightKappa // 海の開始位置
	heightSea     = 4                         // 海の高さ
	countCloud    = 5                         // 雲の個数
	countWave     = 8                         // 波の個数
	offsetMessage = offsetSea + heightSea + 2 // メッセージの開始位置

	tickerInterval = 100 * time.Millisecond
	skyUpdateUnit  = 17
	seaUpdateUnit  = 8
)

// アニメーション構造
var (
	dish         = " ,< = > "
	face         = "(  ' e')"
	body         = "(|   #|"
	foot         = "~~~~~~~~"
	waterSurface = "~"
	cloud        = "---"
	wave         = "^^^^^"
)

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

type Animation struct {
	kappa  *KappaAnim
	sky    *SkySea
	sea    *SkySea
	offset int
	width  int
	msg    string
}

func NewAnimation(msg string) *Animation {
	width, _ := termbox.Size()
	return &Animation{
		kappa:  NewKappaAnim(msg),
		sky:    NewSkySea(heightSky, countCloud),
		sea:    NewSkySea(heightSea, countWave),
		offset: 0,
		width:  width,
		msg:    msg,
	}
}

func NewKappaAnim(msg string) *KappaAnim {
	// 河童のアニメーションリスト
	frames := []kappaFrame{
		{"", "", "", ""},                       // 0
		{"", "", dish, foot},                   // 1
		{"", dish, face, foot},                 // 2
		{dish, face, body, foot},               // 3
		{dish, face + " < " + msg, body, foot}, // 4
	}

	// こんな感じの動き
	actions := []int{0, 0, 0, 1, 2, 3, 3, 3, 3, 3, 4, 4, 4, 4, 3, 3, 2, 1}
	return &KappaAnim{frames: frames, actions: actions, index: 0}
}

// 空と海は構造が同じなので面倒だから一緒くたに定義
func NewSkySea(height, count int) *SkySea {
	offset := make([][]int, height)
	for i := range offset {
		offset[i] = make([]int, count)
	}
	return &SkySea{offset: offset}
}

// 雲と波をランダム生成
func (ss *SkySea) makeRandomOffset(width int) {
	for i := range ss.offset {
		for j := range ss.offset[i] {
			ss.offset[i][j] = rand.Intn(width)
		}
	}
}

func (k *KappaAnim) getFrame() kappaFrame {
	return k.frames[k.actions[k.index]]
}

func (k *KappaAnim) drawKappa(offset int) {
	k.index = (k.index + 1) % len(k.actions)
	frame := k.getFrame()
	drawString(offset, offsetKappa, frame.l1)
	drawString(offset, offsetKappa+1, frame.l2)
	drawString(offset, offsetKappa+2, frame.l3)
	drawString(offset, offsetKappa+3, frame.l4)
}

func (ss *SkySea) drawSky() {
	for i, row := range ss.offset {
		for _, x := range row {
			drawString(x, offsetSky+i, cloud)
		}
	}
}

func (ss *SkySea) drawSea(width int) {
	// 画面幅いっぱいに波線を描画
	surface := strings.Repeat(waterSurface, width)
	drawString(0, offsetSea-2, surface)
	for i, row := range ss.offset {
		for _, x := range row {
			// 水面から１マス下に河童の胴体がくるので、さらにもう１マス下から開始する
			drawString(x, offsetSea+i, wave)
		}
	}
}

// 規定のx, yに文字を表示
func drawString(x, y int, s string) {
	for i, r := range s {
		termbox.SetCell(x+i, y, r, termbox.ColorWhite, termbox.ColorDefault)
	}
}
func (a *Animation) Run() error {
	ticker := time.NewTicker(tickerInterval)
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
			// Enterキーとqキーで停止
			// ctrl+Cも明示的に停止を許容
			if ev.Type == termbox.EventKey && (ev.Key == termbox.KeyEnter || ev.Ch == 'q' || ev.Key == termbox.KeyCtrlC) {
				return nil
			}
		case <-ticker.C:
			if err := a.draw(); err != nil {
				return err
			}
			a.offset = (a.offset + 1) % a.width
		}
	}
}

func (a *Animation) draw() error {
	if err := termbox.Clear(termbox.ColorDefault, termbox.ColorDefault); err != nil {
		return fmt.Errorf("failed to clear screen: %w", err)
	}
	if a.offset%skyUpdateUnit == 0 {
		a.sky.makeRandomOffset(a.width)
	}
	if a.offset%seaUpdateUnit == 0 {
		a.sea.makeRandomOffset(a.width)
	}

	a.sky.drawSky()
	a.sea.drawSea(a.width)
	a.kappa.drawKappa(a.offset)

	drawString(0, offsetMessage, "Press the 'q' key or Enter to stop.")
	return termbox.Flush()
}

func main() {
	msg := flag.String("m", "Hello!", "Define the message that Kappa speaks.")
	flag.Parse()

	if err := termbox.Init(); err != nil {
		fmt.Printf("Failed to initialize termbox: %v\n", err)
		return
	}
	defer termbox.Close()
	animation := NewAnimation(*msg)
	rand.New(rand.NewSource(time.Now().UnixNano()))
	if err := animation.Run(); err != nil {
		fmt.Printf("Animation error: %v\n", err)
	}
}
