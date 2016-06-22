package Logger

import(
	"os"
	"github.com/op/go-logging"
)

var log *logging.Logger
var syslogID = "github.com/andreyevsyukov/coap"

func Init() {

	log = logging.MustGetLogger("coap.default")
	log.ExtraCalldepth++ // Increase +1 to avoid wrapper call from call stack

	//format := logging.MustStringFormatter("%{module} %{shortfile} > %{level:.7s} > %{message}")
	format := logging.MustStringFormatter(syslogID+" %{shortfile} > %{level:.7s} > %{message}")

    //file to stdout
    stdLog          := logging.NewLogBackend(os.Stderr, "", 0)
    stdLogFormatter := logging.NewBackendFormatter(stdLog, format)

    //log to syslog
    syslogLogger, err := logging.NewSyslogBackend(syslogID)
    if err != nil {
    	panic(err)
    }
    syslogFormatter   := logging.NewBackendFormatter(syslogLogger, format)

    //setup logs
    //if config.Env == "prod" {
    if false {
        syslogLeveled := logging.AddModuleLevel(syslogFormatter)
        //log3Leveled.SetLevel(logging.INFO, "")
        syslogLeveled.SetLevel(logging.WARNING, "")
        log.SetBackend(logging.MultiLogger(syslogLeveled))
    } else {
        log.SetBackend(logging.MultiLogger(stdLogFormatter, syslogFormatter))
    }
}

func GetSyslogID() string {
	return syslogID
}

func SetSyslogID(id string) {
	syslogID = id
}

func Debug(v ...interface{}) {
	if log != nil {
		log.Debug(v)
	}
}

func Error(v ...interface{}) {
	if log != nil {
		log.Error(v)
	}
}

func Warning(v ...interface{}) {
	if log != nil {
		log.Warning(v)
	}
}

func Fatal(v ...interface{}) {
	if log != nil {
		log.Fatal(v)
	}
}