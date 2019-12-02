package logger

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lengyizhou/go-library/pool"
)

type fileDriver struct {
	mu *sync.RWMutex

	driver   *log.Logger // use golang default logger now
	lf       *os.File    // 日志文件
	workPool *pool.Pool  // 写日志工作池
	dir      string      // 日志目录
	fn       string      // 日志文件名
	fullname string      // 日志目录+文件名
	lv       int         // 日志等级
	flag     int         // 日志flag
	depth    int         // 日志flag
	st       int         // 日志切分类型
	delay    int         // 当切分类型为delay时必须, 按照天或者小时或者分钟切分
	fm       string      // 当切分类型为delay时必须, 切分的日期格式
	dt       *time.Time  // 当切分类型为delay时必须
	size     int64       // 日志文件大小, 当切分类型为size时必须
	suffix   int         // 当切分类型为size时必须, 日志文件的数量后缀
	fc       int         // 当切分类型为size时必须, 当前目录下已有的日志文件数量

}

var fileSuffix = ".log"

func (fd *fileDriver) mustSplit() bool {
	now := time.Now()
	if fd.st == SplitAsSize {
		if fd.fc >= 1 {
			f, e := os.Stat(fd.fn)
			if e == nil {
				if f.Size() >= fd.size {
					return true
				}
			}
		}
	} else if fd.st == SplitAsDelayDay {
		t, _ := time.Parse(fd.fm, now.Format(fd.fm))
		return t.After((*fd.dt).AddDate(0, 0, fd.delay))
	} else if fd.st == SplitAsDelayHour {
		t, _ := time.Parse(DelayDayFormat+fd.fm, now.Format(DelayDayFormat+fd.fm))
		return t.After((*fd.dt).Add(time.Hour * time.Duration(fd.delay)))
	} else if fd.st == SplitAsDelayMinute {
		t, _ := time.Parse(DelayDayFormat+DelayHourFormat+fd.fm,
			now.Format(DelayDayFormat+DelayHourFormat+fd.fm))
		return t.After((*fd.dt).Add(time.Minute * time.Duration(fd.delay)))
	}

	return false
}
func (fd *fileDriver) Initialize(logDir, logName string, loggerSplit, level, flag int, args ...int) {
	err := os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		panic(err)
	}
	if logName == "" {
		panic("logger file name is empty")
	}

	if loggerSplit == SplitAsSize {
		size := args[0]
		if size <= 0 || size > 500 {
			panic("logger split as size must in 1~500(M)")
		}
		fd.size = int64(size * 1024 * 1024)
		fd.suffix = 0
	} else if loggerSplit == SplitAsDelayDay {
		fd.delay = 1
		fd.fm = DelayDayFormat
	} else if loggerSplit == SplitAsDelayHour {
		fd.delay = 1
		fd.fm = DelayHourFormat
	} else if loggerSplit == SplitAsDelayMinute {
		fd.delay = 1
		fd.fm = DelayMinuteFormat
	} else {
		panic("invalid logger split type")
	}

	if level < LevelInfo || level > LevelError {
		panic("invalid logger level")
	}
	now := time.Now()
	fd.mu = new(sync.RWMutex)

	fd.workPool = pool.NewPool(16, 1024)
	fd.fn = logName
	fd.fullname = ""
	fd.dir = logDir
	fd.lv = level
	fd.flag = flag
	fd.depth = 1
	fd.st = loggerSplit
	fd.dt = &now

	fd.mu.Lock()
	if fd.st == SplitAsSize {
		fd.fc = fd.fileCount()
		fd.fn = filepath.Join(fd.dir, fd.fn)
		for i := 0; i < fd.fc; i++ {
			_fn := fd.fn + "_" + strconv.Itoa(i) + fileSuffix
			_, err = os.Stat(_fn)
			if err != nil {
				if !os.IsNotExist(err) {
					panic(err)
				} else {
					break
				}
			} else {
				fd.suffix += 1
			}
		}
		fd.fullname = fd.fn
		if !fd.mustSplit() {
			err = fd.create()
			if err != nil {
				panic(err)
			}
		} else {
			err = fd.split()
			if err != nil {
				panic(err)
			}
		}
	} else if fd.st == SplitAsDelayDay {
		fd.fullname = filepath.Join(fd.dir, fd.fn+"-"+now.Format(fd.fm)+fileSuffix)
		if !fd.mustSplit() {
			_, err = os.Stat(fd.fullname)
			if err != nil {
				if !os.IsNotExist(err) {
					panic(err)
				}
			}
			err = fd.create()
			if err != nil {
				panic(err)
			}
		} else {
			err = fd.split()
			if err != nil {
				panic(err)
			}
		}
	} else if fd.st == SplitAsDelayHour {
		dir := filepath.Join(fd.dir, now.Format("2006-01-02"))
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			panic(err)
		}
		fd.fullname = filepath.Join(dir, now.Format(fd.fm)+fileSuffix)
		if !fd.mustSplit() {
			_, err = os.Stat(fd.fullname)
			if err != nil {
				if !os.IsNotExist(err) {
					panic(err)
				}
			}
			err = fd.create()
			if err != nil {
				panic(err)
			}
		} else {
			err = fd.split()
			if err != nil {
				panic(err)
			}
		}
	} else {
		dir := filepath.Join(fd.dir, now.Format("2006-01-02"), now.Format("15"))
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			panic(err)
		}
		fd.fullname = filepath.Join(dir, now.Format(fd.fm)+fileSuffix)
		if !fd.mustSplit() {
			_, err = os.Stat(fd.fullname)
			if err != nil {
				if !os.IsNotExist(err) {
					panic(err)
				}
			}
			err = fd.create()
			if err != nil {
				panic(err)
			}
		} else {
			err = fd.split()
			if err != nil {
				panic(err)
			}
		}
	}

	go fd.monitor()

	fd.mu.Unlock()
}

func (fd *fileDriver) fileCount() int {
	count := 0
	files, _ := ioutil.ReadDir(fd.dir)
	for _, f := range files {
		if strings.HasPrefix(f.Name(), fd.fn) {
			count++
		}
	}
	return count

}
func (fd *fileDriver) create() error {
	var err error
	fd.lf, err = os.OpenFile(fd.fullname, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	fd.driver = log.New(fd.lf, "", fd.flag)
	return nil
}

func (fd *fileDriver) split() error {
	var err error
	if fd.lf != nil {
		fd.lf.Close()
	}
	if fd.st == SplitAsSize {
		fd.suffix = fd.suffix + 1
		if fd.lf != nil {
			fd.lf.Close()
		}

		lfBak := fd.fn + "_" + strconv.Itoa(fd.suffix) + fileSuffix
		_, err = os.Stat(lfBak)
		if err == nil {
			os.Remove(lfBak)
		}
		os.Rename(fd.fn, lfBak)
		err = fd.create()
		if err != nil {
			return err
		}
	} else if fd.st == SplitAsDelayDay {
		now := time.Now()
		fd.fullname = filepath.Join(fd.dir, now.Format(fd.fm)+fileSuffix)
		err = fd.create()
		if err != nil {
			return err
		}
		fd.dt = &now
	} else if fd.st == SplitAsDelayHour {
		now := time.Now()
		dir := filepath.Join(fd.dir, now.Format("2006-01-02"))
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
		fd.fullname = filepath.Join(dir, now.Format(fd.fm)+fileSuffix)
		err = fd.create()
		if err != nil {
			return err
		}
		fd.dt = &now
	} else {
		now := time.Now()
		dir := filepath.Join(fd.dir, now.Format("2006-01-02"), now.Format("15"))
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
		fd.fullname = filepath.Join(dir, now.Format(fd.fm)+fileSuffix)
		err = fd.create()
		if err != nil {
			return err
		}
		fd.dt = &now
	}
	return nil
}

func (fd *fileDriver) monitor() {
	defer func() {
		if err := recover(); err != nil {
			fd.driver.Printf("Logger's monitor() catch panic: %v\n", err)
		}
	}()

	if fd.st == SplitAsSize {
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ticker.C:
				fd.check()
			}
		}
	} else {
		ticker := time.NewTicker(60 * time.Second)
		for {
			select {
			case <-ticker.C:
				fd.check()
			}
		}
	}
}

func (fd *fileDriver) check() {
	defer func() {
		if err := recover(); err != nil {
			fd.driver.Printf("Logger's check() catch panic: %v\n", err)
		}
	}()

	if fd.mustSplit() {
		fd.mu.Lock()
		fd.split()
		fd.mu.Unlock()
	}
}

func (fd *fileDriver) Error(format string, v ...interface{}) {
	if LevelError > fd.lv {
		return
	}
	fd.workPool.JobQueue <- func() {
		fd.driver.SetPrefix("[Error]")
		fd.driver.Output(fd.depth, fmt.Sprintf(format, v...))
	}
}

func (fd *fileDriver) Debug(format string, v ...interface{}) {
	if LevelDebug > fd.lv {
		return
	}
	fd.workPool.JobQueue <- func() {
		fd.driver.SetPrefix("[Debug]")
		fd.driver.Output(fd.depth, fmt.Sprintf(format, v...))
	}
}

func (fd *fileDriver) Warning(format string, v ...interface{}) {
	if LevelWarning > fd.lv {
		return
	}
	fd.workPool.JobQueue <- func() {
		fd.driver.SetPrefix("[Warning]")
		fd.driver.Output(fd.depth, fmt.Sprintf(format, v...))
	}
}

func (fd *fileDriver) Info(format string, v ...interface{}) {
	if LevelInfo > fd.lv {
		return
	}
	fd.workPool.JobQueue <- func() {
		fd.driver.SetPrefix("[Info]")
		fd.driver.Output(fd.depth, fmt.Sprintf(format, v...))
	}
}

func (fd *fileDriver) Close() error {
	if fd.lf != nil {
		return fd.lf.Close()
	}
	fd.workPool.Release()
	return nil
}

func init() {
	regDrivers("file", &fileDriver{})
}
