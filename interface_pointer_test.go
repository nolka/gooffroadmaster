package main

import (
	"testing"
)

type SomeInterface interface {
	DoSomething()
}

type SomethingWorker struct{

}

func (w *SomethingWorker) DoSomething() {

}

func GetWorker() SomeInterface {
	w := new(SomethingWorker)
	return w
}

func TestInterfacePointer(t *testing.T) {


}
