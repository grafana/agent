package armclient

import (
	"github.com/remeh/sizedwaitgroup"
)

type (
	InterfaceIterator struct {
		list []interface{}

		concurrency int
	}
)

func NewInterfaceIterator() *InterfaceIterator {
	i := InterfaceIterator{}
	i.concurrency = IteratorDefaultConcurrency
	return &i
}

func (i *InterfaceIterator) SetList(list ...interface{}) *InterfaceIterator {
	i.list = list
	return i
}

func (i *InterfaceIterator) GetList() []interface{} {
	return i.list
}

func (i *InterfaceIterator) SetConcurrency(concurrency int) *InterfaceIterator {
	i.concurrency = concurrency
	return i
}

func (i *InterfaceIterator) ForEach(callback func(object interface{})) error {
	for _, subscription := range i.list {
		callback(subscription)
	}
	return nil
}

func (i *InterfaceIterator) ForEachAsync(callback func(object interface{})) error {
	wg := sizedwaitgroup.New(i.concurrency)

	for _, object := range i.list {
		wg.Add()

		go func(object interface{}) {
			defer wg.Done()
			callback(object)
		}(object)
	}

	wg.Wait()
	return nil
}
