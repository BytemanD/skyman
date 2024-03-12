package utility

import (
	"fmt"
	"reflect"
	"sync"
)

func GoroutineMap(function func(item interface{}), items interface{}) error {
	value := reflect.ValueOf(items)
	if value.Kind() != reflect.Slice && value.Kind() != reflect.Array {
		return fmt.Errorf("items must be slice or array")
	}
	wg := sync.WaitGroup{}
	wg.Add(value.Len())
	for i := 0; i < value.Len(); i++ {
		go func(o interface{}, wg *sync.WaitGroup) {
			defer wg.Done()
			function(o)
		}(value.Index(i).Interface(), &wg)
	}
	wg.Wait()
	return nil
}
