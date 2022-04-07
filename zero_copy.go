package zmq4

/*
#include <zmq.h>
#include <stdint.h>
#include <stdlib.h>
#include "zmq4.h"

   // extern void go_callback(int foo);

void my_free (void *data, void *hint)
{
   printf("======== in callback ========\n");
   printf("data: %s\n", data);
   printf("hint: %s\n", hint);
   printf("========== ==========\n");

   // go_callback(42);
}

*/
import "C"
import (
	"fmt"
	"unsafe"
)

// //export go_callback
// func go_callback(foo C.int) {
// 	fmt.Printf("C.int: %d\n", foo)
// }

func checkError(ret C.int) error {
	if ret == 0 {
		return nil
	}

	c_char := C.zmq_strerror(ret)
	return fmt.Errorf("zmq error: %v (%d)", C.GoString(c_char), int(ret))
}

func PrototypeSend(socket *Socket, data []byte) error {
	if !socket.opened {
		return fmt.Errorf("socket not opened")
	}

	var ret C.int
	var c_msg C.zmq_msg_t
	var cb *C.zmq_free_fn

	c_buf := unsafe.Pointer(&data[0])
	cb = (*C.zmq_free_fn)(C.my_free)
	ret = C.zmq_msg_init_data((*C.zmq_msg_t)(unsafe.Pointer(&c_msg)), c_buf, C.size_t(len(data)), cb, nil)
	if err := checkError(ret); err != nil {
		return err
	}

	ret = C.zmq_sendmsg(socket.soc, (*C.zmq_msg_t)(unsafe.Pointer(&c_msg)), 0)
	if ret < 0 {
		return fmt.Errorf("zmq send msg, check errno?? : %d", ret)
	}

	return nil
}

// // SendNC ...
// func (self *Socket) SendNC(data []byte) error {

// }
