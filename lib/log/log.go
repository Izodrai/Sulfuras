package log

import (
	"fmt"
	"github.com/cheggaaa/pb"
	"log"
	"os"
	"syscall"
	"../config"
)

/*
 * Const for colors
 */
const (
	STOP  = "\x1b[0m"
	WHITE = "\x1b[37;1m"
	RED   = "\x1b[31;1m"
	GREEN = "\x1b[32;1m"
	CYAN  = "\x1b[36m"

	YELLOW     = "\x1b[33;1m"
	BRIGHTBLUE = "\x1b[34;1m"
	PURPLE     = "\x1b[35;1m"
	BRIGHTCYAN = "\x1b[36;1m"
)

/*
 * Var for display
 */
var logDisplay *os.File
var loggerInfo *log.Logger
var loggerError *log.Logger
var loggerFatalError *log.Logger
var loggerWarning *log.Logger
var loggerDebug *log.Logger
var debug bool
var Fatal bool

/*
 * Func for init log
 */
func InitLog(d bool, config config.Config) error {

	debug = d

	fLog, err := os.OpenFile(config.LogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
  syscall.Dup2(int(fLog.Fd()), 1) /* -- stdout */

	logDisplay = os.Stdout

	logMode := log.Ldate | log.Ltime

	loggerInfo = log.New(logDisplay, "INFO    ", logMode)
	loggerError = log.New(logDisplay, RED+"ERROR   ", logMode)
	loggerFatalError = log.New(logDisplay, RED+"FATAL   ", logMode)
	loggerWarning = log.New(logDisplay, YELLOW+"WARNING ", logMode)
	loggerDebug = log.New(logDisplay, CYAN+"DEBUG   ", logMode)

	return nil
}

func FatalError(v ...interface{}) {
	loggerFatalError.Println(RED + fmt.Sprint(v...) + STOP)
	// 	os.Exit(0)
}
func Error(v ...interface{}) {
	loggerError.Println(RED + fmt.Sprint(v...) + STOP)
}
func Warning(v ...interface{}) {
	loggerWarning.Println(YELLOW + fmt.Sprint(v...) + STOP)
}
func Debug(v ...interface{}) {
	if debug {
		loggerDebug.Println(CYAN + fmt.Sprint(v...) + STOP)
	}
}
func Info(v ...interface{}) {
	loggerInfo.Println(fmt.Sprint(v...))
}
func GreenInfo(v ...interface{}) {
	loggerInfo.Println(GREEN + fmt.Sprint(v...) + STOP)
}
func BlueInfo(v ...interface{}) {
	loggerInfo.Println(BRIGHTBLUE + fmt.Sprint(v...) + STOP)
}
func YellowInfo(v ...interface{}) {
	loggerInfo.Println(YELLOW + fmt.Sprint(v...) + STOP)
}
func PurpleInfo(v ...interface{}) {
	loggerInfo.Println(PURPLE + fmt.Sprint(v...) + STOP)
}
func WhiteInfo(v ...interface{}) {
	loggerInfo.Println(WHITE + fmt.Sprint(v...) + STOP)
}
func CyanInfo(v ...interface{}) {
	loggerInfo.Println(BRIGHTCYAN + fmt.Sprint(v...) + STOP)
}

func InitBar(l int, display bool) *pb.ProgressBar {

	bar := pb.StartNew(l)
	bar.ShowPercent = display
	bar.ShowCounters = display
	bar.ShowSpeed = display
	bar.ShowTimeLeft = display
	bar.ShowBar = display
	bar.ShowFinalTime = display

	bar.SetWidth(80)
	bar.SetMaxWidth(80)

	return bar
}
