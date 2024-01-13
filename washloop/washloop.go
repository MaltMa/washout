// Copyright © 2018 shoarai

package washloop

import (
	"sync"
	"time"

	"github.com/shoarai/washout"
)

// A Motion is accelerations and angular velocities in 3D axis.运动是3D轴上的加速度和角速度。
type Motion struct {
	Acceleration    washout.Vector
	AngularVelocity washout.Vector
}

// A WashoutInterface is a interface of washout.WashoutInterface是一个washout接口。
type WashoutInterface interface {
	Filter(
		accelerationX, accelerationY, accelerationZ,
		angularVelocityX, angularVelocityY, angularVelocityZ float64) washout.Position
}

// A WashoutLoop is a loop for process of washout.WashoutLoop是用于清洗过程的循环。
type WashoutLoop struct {
	interval uint

	washout  WashoutInterface
	motion   Motion
	position washout.Position

	stopCh chan struct{}

	motionMutex   *sync.Mutex
	positionMutex *sync.Mutex
}

// NewWashLoop creates new washout loop.NewWashLoop创建新的洗出循环。
func NewWashLoop(washout WashoutInterface, interval uint) *WashoutLoop {
	w := WashoutLoop{}
	w.stopCh = make(chan struct{})
	w.interval = interval
	w.init(washout)
	return &w
}

// Start starts a loop of process.Start启动进程的循环。
func (w *WashoutLoop) Start() {
	interval := time.Duration(w.interval) * time.Millisecond
	ticker := time.NewTicker(interval)
	w.filter()
	for {
		select {
		case <-ticker.C:
			w.filter()
		case <-w.stopCh:
			ticker.Stop()
			return
		}
	}
}

// Stop stops a loop of process.Stop停止进程的循环。
func (w *WashoutLoop) Stop() {
	close(w.stopCh)
}

func (w *WashoutLoop) filter() {
	motion := w.getMotion()
	position := w.washout.Filter(
		motion.Acceleration.X,
		motion.Acceleration.Y,
		motion.Acceleration.Z,
		motion.AngularVelocity.X,
		motion.AngularVelocity.Y,
		motion.AngularVelocity.Z,
	)
	w.setPosition(position)
}

func (w *WashoutLoop) init(washout WashoutInterface) {
	w.washout = washout
	w.motionMutex = new(sync.Mutex)
	w.positionMutex = new(sync.Mutex)
}

// SetMotion sets a motion used as input of washout.SetMotion设置用作冲洗输入的运动。
func (w *WashoutLoop) SetMotion(motion Motion) {
	w.motionMutex.Lock()
	defer w.motionMutex.Unlock()

	w.motion = motion
}

func (w *WashoutLoop) getMotion() Motion {
	w.motionMutex.Lock()
	defer w.motionMutex.Unlock()

	return w.motion
}

func (w *WashoutLoop) setPosition(position washout.Position) {
	w.positionMutex.Lock()
	defer w.positionMutex.Unlock()

	w.position = position
}

// GetPosition gets a position as output of washout.GetPosition获取一个位置作为washout的输出。
func (w *WashoutLoop) GetPosition() washout.Position {
	w.positionMutex.Lock()
	defer w.positionMutex.Unlock()

	return w.position
}
