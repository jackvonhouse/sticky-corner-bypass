package winapi

import (
	"fmt"
	"syscall"
	"unsafe"
)

type CursorPoint struct {
	X, Y int32
}

type rectangle struct {
	Left, Top, Right, Bottom int32
}

type monitorInfo struct {
	CountOfByteSize uint32
	Monitor         rectangle
	Work            rectangle
	Flags           uint32
}

type Monitor struct {
	rectangle
	Main bool
}

type Windows struct {
	user32Dll *syscall.LazyDLL

	getCursorPosFunction   *syscall.LazyProc
	getMonitorInfoFunction *syscall.LazyProc
	setCursorPosFunction   *syscall.LazyProc

	monitors []*Monitor
}

func New() *Windows {
	dll := syscall.NewLazyDLL("user32.dll")

	getCursorPosFunction := dll.NewProc("GetCursorPos")
	getMonitorInfoFunction := dll.NewProc("GetMonitorInfoW")
	enumDisplayMonitorsFunction := dll.NewProc("EnumDisplayMonitors")
	setCursorPosFunction := dll.NewProc("SetCursorPos")

	monitors := make([]*Monitor, 0)

	getMonitorInfoCallback := func(
		hMonitor uintptr,
		hdc uintptr,
		rect *rectangle,
		lparam uintptr,
	) uintptr {

		monitorInfo := monitorInfo{}
		monitorInfo.CountOfByteSize = uint32(unsafe.Sizeof(monitorInfo))

		getMonitorInfoFunction.Call(hMonitor, uintptr(unsafe.Pointer(&monitorInfo)))

		monitor := Monitor{
			Main:      monitorInfo.Flags == 1,
			rectangle: monitorInfo.Monitor,
		}

		monitors = append(monitors, &monitor)

		return 1
	}

	cb := syscall.NewCallback(getMonitorInfoCallback)

	enumDisplayMonitorsFunction.Call(0, 0, cb, 0)

	return &Windows{
		user32Dll: dll,

		getCursorPosFunction:   getCursorPosFunction,
		getMonitorInfoFunction: getMonitorInfoFunction,
		setCursorPosFunction:   setCursorPosFunction,

		monitors: monitors,
	}
}

func (w *Windows) GetCursorPosition() CursorPoint {
	point := CursorPoint{}

	syscall.SyscallN(
		w.getCursorPosFunction.Addr(),
		uintptr(unsafe.Pointer(&point)),
	)

	return point
}

func (w *Windows) SetCursorPosition(pos CursorPoint) {
	w.setCursorPosFunction.Call(uintptr(pos.X), uintptr(pos.Y))
}

func (w *Windows) GetMonitor(pos CursorPoint) (*Monitor, error) {
	for _, monitor := range w.monitors {
		if pos.X >= monitor.Left && pos.X < monitor.Right {
			return monitor, nil
		}
	}

	return nil, fmt.Errorf("cant get current monitor by passed position: %v", pos)
}

func (w *Windows) CanSetCursorPosition(pos CursorPoint) bool {
	for _, monitor := range w.monitors {
		if pos.X > monitor.Left && pos.X < monitor.Right {
			return true
		}
	}

	return false
}

func (w *Windows) TrySetCursorPosition(pos CursorPoint) {
	if w.CanSetCursorPosition(pos) {
		w.SetCursorPosition(pos)
	}
}

func (w *Windows) Proccess() error {
	for {
		cursor := w.GetCursorPosition()
		currentMonitor, err := w.GetMonitor(cursor)
		if err != nil {
			return err
		}

		isTop := cursor.Y == currentMonitor.Top
		isBottom := cursor.Y == currentMonitor.Bottom-1

		if !isTop && !isBottom {
			continue
		}

		newCursorPosition := CursorPoint{
			X: cursor.X,
			Y: cursor.Y,
		}

		if cursor.X == currentMonitor.Left {
			newCursorPosition.X -= 2
		} else if cursor.X == currentMonitor.Right-1 {
			newCursorPosition.X += 2
		} else {
			continue
		}

		if !w.CanSetCursorPosition(newCursorPosition) {
			continue
		}

		newMonitor, err := w.GetMonitor(newCursorPosition)
		if err != nil {
			return err
		}

		if isTop {
			newCursorPosition.Y = newMonitor.Top
		} else if isBottom {
			newCursorPosition.Y = newMonitor.Bottom - 1
		}

		w.SetCursorPosition(newCursorPosition)
	}
}
