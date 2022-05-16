package zmq4

/*
#include <zmq.h>
#include <stdint.h>
#include <stdlib.h>
#include "zmq4.h"
#include "zero_copy.h"
*/
import "C"
import (
	"fmt"
	"sync"
	"syscall"
	"unsafe"
)

var Data []byte

func checkError(ret C.int) error {
	if ret == 0 {
		return nil
	}

	c_char := C.zmq_strerror(ret)
	return fmt.Errorf("zmq error: %v (%d)", C.GoString(c_char), int(ret))
}

type ZCMessage *C.zmq_msg_t

// type ZCMessage struct {
// 	data  []byte
// 	c_msg *C.zmq_msg_t
// }

// Initialize
// func (self *ZCMessage) Initialize(data []byte) error {
// 	var c_ret C.int
// 	var c_msg C.zmq_msg_t
// 	var c_callback *C.zmq_free_fn = (*C.zmq_free_fn)(C.my_free)

// 	c_buf := unsafe.Pointer(&data[0])
// 	c_msg_ptr := (*C.zmq_msg_t)(unsafe.Pointer(&c_msg))
// 	c_size := C.size_t(len(data))

// 	c_ret = C.zmq_msg_init_data(c_msg_ptr, c_buf, c_size, c_callback, nil)
// 	if err := checkError(c_ret); err != nil {
// 		return err
// 	}

// 	self.data = data
// 	self.c_msg = c_msg_ptr
// 	return nil
// }

var hintLock sync.Mutex
var hintStore map[unsafe.Pointer]hint = map[unsafe.Pointer]hint{}

type hint struct {
	soc    unsafe.Pointer
	length int
}

// InitializeZeroCopy
func (self *Socket) ZCInit() error {
	if self.zcpipe != nil {
		return fmt.Errorf("could not initialize, pipe alread exists")
	}
	self.zcpipe = make(chan []byte)

	var ptr unsafe.Pointer = C.malloc(C.size_t(1))
	if ptr == nil {
		panic("can't allocate 'cgo-pointer hack index pointer': ptr == nil")
	}

	self.zcID = ptr

	fmt.Printf("ZMQ: Socket pointer: %v\n", self.zcID)

	return nil
}

// Pipe
func (self *Socket) ZCPipe() (<-chan []byte, error) {
	if self.zcpipe == nil {
		return nil, fmt.Errorf("pipe not initialized, call InitializeZeroCopy")
	}

	return self.zcpipe, nil
}

// InitializeMessage
func (self *Socket) ZCInitMessage(data []byte) (ZCMessage, error) {
	var c_ret C.int
	var c_msg C.zmq_msg_t
	var c_callback *C.zmq_free_fn = (*C.zmq_free_fn)(C.my_free)

	c_buf := unsafe.Pointer(&data[0])
	c_msg_ptr := (*C.zmq_msg_t)(unsafe.Pointer(&c_msg))
	c_size := C.size_t(len(data))

	hint := hint{unsafe.Pointer(self), cap(data)}
	c_hint := unsafe.Pointer(C.malloc(C.size_t(1)))

	c_ret = C.zmq_msg_init_data(c_msg_ptr, c_buf, c_size, c_callback, c_hint)
	if err := checkError(c_ret); err != nil {
		return c_msg_ptr, err
	}

	hintLock.Lock()
	hintStore[c_hint] = hint
	hintLock.Unlock()

	return c_msg_ptr, nil
}

// SendZeroCopyMessage ...
func (self *Socket) ZCSend(message ZCMessage) error {
	var c_ret C.int
	c_ret = C.zmq_sendmsg(self.soc, message, 0)
	if c_ret < 0 {
		errno := Errno(C.zmq_errno())
		if errno == AsErrno(Errno(syscall.EINTR)) {
			// fmt.Println("========================================")
			// fmt.Println("===========     RETRYING     ===========")
			// fmt.Println("========================================")
			c_ret = C.zmq_sendmsg(self.soc, message, 0)
		}

		if c_ret < 0 {
			return fmt.Errorf("zmq send msg, check errno?? : %d", c_ret)
		}

	}

	return nil
}

//export go_callback
func go_callback(c_data unsafe.Pointer, c_hint unsafe.Pointer) {
	hintLock.Lock()
	hint := hintStore[c_hint]
	delete(hintStore, c_hint)
	hintLock.Unlock()

	soc := (*Socket)(hint.soc)
	buf := unsafe.Slice((*byte)(c_data), hint.length)

	soc.zcpipe <- buf

	// fmt.Printf("ZMQ: hint: %d\n", hint.length)

	// fmt.Printf("ZMQ: Socket? %v\n", soc)

	// fmt.Printf("ZMQ: Data: %p %v\n", buf, buf)

	// fmt.Printf("ZMQ: data: %v, hint: %v\n", data, hint)
	// buf := unsafe.Slice((*byte)(data), 5)
	// fmt.Printf("ZMQ: callback: %p\n", buf)
	// Data = buf

	// fmt.Println("==================================================")
	// fmt.Println("==================================================")
	// var v interface{} = store[hint]
	// soc := v.(*Socket)
	// fmt.Printf("ZMQ: Socket? %T %v \n", soc, soc)

	// soc.zcpipe <- buf
	// fmt.Printf("ZMQ: After pipe\n")
}

// func PrototypeSend(socket *Socket, data []byte) error {
// 	if !socket.opened {
// 		return fmt.Errorf("socket not opened")
// 	}

// 	var c_ret C.int
// 	var c_msg C.zmq_msg_t
// 	var c_callback *C.zmq_free_fn = (*C.zmq_free_fn)(C.my_free)

// 	c_buf := unsafe.Pointer(&data[0])
// 	c_msg_ptr := (*C.zmq_msg_t)(unsafe.Pointer(&c_msg))
// 	c_size := C.size_t(len(data))

// 	c_ret = C.zmq_msg_init_data(c_msg_ptr, c_buf, c_size, c_callback, nil)
// 	if err := checkError(c_ret); err != nil {
// 		return err
// 	}

// 	// ==================================================
// 	//
// 	// imagine long time here...
// 	//
// 	// ==================================================

// 	c_ret = C.zmq_sendmsg(socket.soc, c_msg_ptr, 0)
// 	if c_ret < 0 {
// 		return fmt.Errorf("zmq send msg, check errno?? : %d", c_ret)
// 	}

// 	return nil
// }
