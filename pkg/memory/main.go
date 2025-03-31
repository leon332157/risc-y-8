package memory

import (
	// "fmt"

	"github.com/leon332157/risc-y-8/pkg/types"
)

var mtoexe types.MemToExe

func MemoryStageToWriteBack(etom types.ExeToMem, cache CacheType) types.MemToWB{

	mtowb := etom.MemToWB

	if etom.IsLoad {

		cache_val := cache.Read(etom.Address)
		mtoexe.Address = etom.Address
		mtowb.RegVal = cache_val

	} else if !etom.IsLoad && !etom.IsControl && !etom.IsALU {

		cache.Write(etom.Address, etom.Data)
		mtoexe.Address = etom.Address
		return types.MemToWB{}
			
	}

	return mtowb
}

func MemoryStageToExecute() types.MemToExe {

	return mtoexe

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
