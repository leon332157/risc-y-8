package memory

import (
	// "fmt"

	"github.com/leon332157/risc-y-8/pkg/types"
)

func MemoryStage(msi types.MemoryStageInput, cache CacheType) types.WriteBackStageInput{

	wbsi := msi.WriteBackStageInput

	if msi.IsLoad {

		cache_val := cache.Read(msi.Address)
		wbsi.RegVal = cache_val

	} else if !msi.IsLoad && !msi.IsControl && !msi.IsALU {

		cache.Write(msi.Address, msi.Data)
		return types.WriteBackStageInput{}
			
	}

	return wbsi

}

// func main() {
// 	fmt.Println("hello")
// }

// func main() {
// 	var mem Memory
// 	ram := mem.CreateRAM(256, 4, 32, 3)
// 	ram.Write(10, &RAMValue{value: 123})
// 	for i := 0; i < 10; i += 1 {

// 		val := mem.Read(10, false)

// 		if val == nil {
// 			fmt.Println("Read: ", "WAIT")
// 		} else {
// 			fmt.Println("Read: ", val.value) // should print "Read: 123"
// 		}
// 	}

// 	var cache Cache
// 	c := cache.CreateDefault(ram)
// 	PrintCache(c)

// 	c.Search(10)
// 	PrintCache(c)

// 	c.Search(32)
// 	PrintCache(c)
// }
