# Go Memory Cache

No periodically scheduler is running without your actions.
It times the cleanup to the next expiry.

### Usage

```go
import(
	"github.com/ProxeusApp/memcache"
	"fmt"
	"time"
)

func main(){
	c := cache.New(3 * time.Second)
	//or extend expiry on get 'true' | 'false'
	c = cache.NewExtendExpiryOnGet(3 * time.Second, true)

	c.Put("myKey", "my Value")
	c.PutWithOtherExpiry("myKey", "my Value", 6*time.Second)
	c.Put(123, 456)
	c.Put(1.456, 7.89)
	c.OnExpired = func(key interface{}, val interface{}){
		fmt.Println("on expired", key, val)
	}
	time.Sleep(10*time.Second)

	var myVal string
	err := c.Get("myKey", &myVal)
	if err != nil {
	    panic(err)
	}
	fmt.Println(myVal)

}
```