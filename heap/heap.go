// Package heap provides min-heap operations in a mostly generic way.
// Most of the implementation is taken from Go's built-in container/heap
// but re-organized to fit a generic pattern that avoids type assertions
// and also doesn't require implementing sort.Interface and list operations.
package heap

import "fmt"
import "reflect"

/* Private Heap Implementation */

type heap struct {
	Data []reflect.Value
	LessImpl reflect.Value
}

func (h *heap) less(i, j int) bool {
	res := h.LessImpl.Call([]reflect.Value{h.Data[i], h.Data[j]})
	return res[0].Bool()
}

func (h *heap) swap(i, j int) {
	h.Data[i], h.Data[j] = h.Data[j], h.Data[i]
}

func (h *heap) Push(in []reflect.Value) []reflect.Value {
	h.Data = append(h.Data, in[0])
	h.up(len(h.Data)-1)
	return nil
}

func (h *heap) Pop(in []reflect.Value) []reflect.Value {
	n := len(h.Data) - 1
	h.swap(0, n)
	h.down(0, n)

	out := h.Data[n]
	h.Data = h.Data[:n]
	return []reflect.Value{out}
}

func (h *heap) Remove(in []reflect.Value) []reflect.Value {
	n := len(h.Data) - 1
	i := in[0].Interface().(int)
	if n != i {
		h.swap(i, n)
		h.down(i, n)
		h.up(i)
	}

	out := h.Data[n]
	h.Data = h.Data[:n]
	return []reflect.Value{out}
}

func (h *heap) up(j int) {
	for {
		i := (j - 1) / 2 // parent
		if i == j || !h.less(j, i) {
			break
		}
		h.swap(i, j)
		j = i
	}
}

func (h *heap) down(i, n int) {
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && !h.less(j1, j2) {
			j = j2 // = 2*i + 2  // right child
		}
		if !h.less(j, i) {
			break
		}
		h.swap(i, j)
		i = j
	}
}

/* Public List Interface */

type GenericHeap struct {
	Heap *heap
}

func (h *GenericHeap) Len() int {
	return len(h.Heap.Data)
}

func getGenericFunc(l reflect.Value, name string)  (func([]reflect.Value) []reflect.Value) {
	return l.MethodByName(name).Interface().(func([]reflect.Value) []reflect.Value)
}

func Init(h interface{}) {
	impl := new(heap)
	implValue := reflect.ValueOf(impl)

	obj := reflect.ValueOf(h).Elem()
	obj.FieldByName("Heap").Set(implValue)

	for _, fieldName := range []string{"Push", "Pop", "Remove"} {
		orig := obj.FieldByName(fieldName)
		if !orig.IsValid() {
			panic(fmt.Sprintf("Expected a definition for %v on GenericHeap", fieldName))
		}
		generic := getGenericFunc(implValue, fieldName)
		orig.Set(reflect.MakeFunc(orig.Type(), generic))
	}

	impl.LessImpl = obj.Addr().MethodByName("Less")
	if !impl.LessImpl.IsValid() {
		panic("Expected implemention of func (*YourHeap) Less(a, b YourType) int")
	}
}
