# go-library

## install
	go get -u -v github.com/lengyizhou/go-library

## configer

### usage

	import "github.com/lengyizhou/go-library/configer"
    
    var conf = configer.New("json", "./config.json")
    conf.Set("key", "value") 
    conf.Get("host") 
    conf.Get("http.host") 
    conf.Int("http.port") 
    conf.String("http.host") 
    conf.DefaultString("http.host", "default") 
    conf.DefaultInt("http.port", 8080) 

### desc
    add functions by self, like: Configer.Bool(), Configer.Int64(), ... and so on.