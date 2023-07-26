package logger

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logConf = Logconfig{
		// 默认info级别
		Level:      "info",
		TimeSep:    "20060101",
		Maxsize:    30,
		Maxage:     30,
		Maxbackups: 30,
		Compress:   true,
	}
	log0 *zap.Logger
)

func init() {
	//解析日志配置
	parseLogConf()
	//配置日志组件
	configLog()

	// 定时按日期切割日志
	//crontab := cron.New(cron.WithSeconds())
	//_, err := crontab.AddFunc("0 0 0 * * ?", func() {
	//	configLog()
	//})
	//if err != nil {
	//	Error("日志定时切割任务创建异常, error:%+v", err)
	//}
	//crontab.Start()
}

// Info info级别日志记录
func Info(format string, args ...any) {
	log.Printf(format+"\n", args...)
	log0.Info(fmt.Sprintf(format, args...))
}

// Warn warn级别日志记录
func Warn(format string, args ...any) {
	log.Printf(format+"\n", args...)
	log0.Warn(fmt.Sprintf(format, args...))
}

// Debug debug级别日志记录
func Debug(format string, args ...any) {
	log.Printf(format+"\n", args...)
	log0.Debug(fmt.Sprintf(format, args...))
}

// Error error级别日志记录
func Error(format string, args ...any) {
	log.Printf(format+"\n", args...)
	log0.Error(fmt.Sprintf(format, args...))
}

// Fatal fatal级别日志记录
func Fatal(format string, args ...any) {
	log.Printf(format+"\n", args...)
	log0.Fatal(fmt.Sprintf(format, args...))
}

// DPanic dPanic级别日志记录
func DPanic(format string, args ...any) {
	log.Printf(format+"\n", args...)
	log0.DPanic(fmt.Sprintf(format, args...))
}

// Panic panic级别日志记录
func Panic(format string, args ...any) {
	log.Printf(format+"\n", args...)
	log0.Panic(fmt.Sprintf(format, args...))
}

// 日志配置
func parseLogConf() {
	vp := viper.New()
	vp.AddConfigPath("config")
	vp.SetConfigName("config.yml")
	vp.SetConfigType("yml")
	err := vp.ReadInConfig()
	if err != nil {
		vp.SetConfigName("config")
		err := vp.ReadInConfig()
		if err != nil {
			fmt.Printf("读取日志配置异常,error:%v\n", err)
			fmt.Printf("将使用默认配置\n")
		}
	}
	err = vp.UnmarshalKey("logconfig", &logConf)
	if err != nil {
		fmt.Printf("解析日志配置异常,error:%v\n", err)
		fmt.Printf("将使用默认配置\n")
	}
}

// Logconfig 日志配置信息
type Logconfig struct {
	Timeformat string `json:"timeformat" yaml:"timeformat"`
	// Level 最低日志等级，DEBUG<INFO<WARN<ERROR<FATAL 例如：info-->收集info等级以上的日志
	Level string `json:"level" yaml:"level"`
	// TimeSep 日志按配置的时间分目录，如按年月日20060101分为一个目录，而第二天20060102会分为第二个目录，同理也可以按年份、月份等分目录
	TimeSep string `json:"timesep" yaml:"timesep"`
	// MaxSize 进行切割之前，日志文件的最大大小(MB为单位)，默认为100MB
	Maxsize int `json:"maxsize" yaml:"maxsize"`
	// MaxAge 是根据文件名中编码的时间戳保留旧日志文件的最大天数。
	Maxage int `json:"maxage" yaml:"maxage"`
	// MaxBackups 是要保留的旧日志文件的最大数量。默认是保留所有旧的日志文件（尽管 MaxAge 可能仍会导致它们被删除。）
	Maxbackups int  `json:"maxbackups" yaml:"maxbackups"`
	Compress   bool `json:"compress" yaml:"compress"`
}

// 负责设置 encoding 的日志格式
func getEncoder() zapcore.Encoder {
	// 获取一个指定的的EncoderConfig，进行自定义
	encodeConfig := zap.NewProductionEncoderConfig()
	// 设置每个日志条目使用的键。如果有任何键为空，则省略该条目的部分。
	// 序列化时间。eg: 2022-09-01T19:11:35.921+0800
	encodeConfig.EncodeTime = formatEncodeTime
	// "time":"2022-09-01T19:11:35.921+0800"
	encodeConfig.TimeKey = "time"
	// 将Level序列化为全大写字符串。例如，将info level序列化为INFO。
	encodeConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	// 以 package/file:行 的格式 序列化调用程序，从完整路径中删除除最后一个目录外的所有目录。
	encodeConfig.EncodeCaller = zapcore.ShortCallerEncoder
	return zapcore.NewJSONEncoder(encodeConfig)
}

// 格式化日期
func formatEncodeTime(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

// 负责日志写入的位置
func getLogWriter(filename string, maxsize, maxBackup, maxAge int) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filename,  // 文件位置
		MaxSize:    maxsize,   // 进行切割之前,日志文件的最大大小(MB为单位)
		MaxAge:     maxAge,    // 保留旧文件的最大天数
		MaxBackups: maxBackup, // 保留旧文件的最大个数
		Compress:   true,      // 是否压缩/归档旧文件
	}
	// AddSync 将 io.Writer 转换为 WriteSyncer。
	// 它试图变得智能：如果 io.Writer 的具体类型实现了 WriteSyncer，我们将使用现有的 Sync 方法。
	// 如果没有，我们将添加一个无操作同步。
	return zapcore.AddSync(lumberJackLogger)
}

// InitLogger 初始化Logger
func configLog() {
	logConf.TimeSep = fmt.Sprintf("var/%s/all.log", time.Now().Format("20060102"))
	// 获取日志写入位置
	writeSyncer := getLogWriter(logConf.TimeSep, logConf.Maxsize, logConf.Maxbackups, logConf.Maxage)
	// 获取日志编码格式
	encoder := getEncoder()

	// 获取日志最低等级，即>=该等级，才会被写入。
	var l = new(zapcore.Level)
	if err := l.UnmarshalText([]byte(logConf.Level)); err != nil {
		log.Printf("日志级别配置刷新异常, error:%+v\n", err)
	}

	// 创建一个将日志写入 WriteSyncer 的核心。
	core := zapcore.NewCore(encoder, writeSyncer, l)
	log0 = zap.New(core, zap.AddCaller())

	// 替换zap包中全局的logger实例，后续在其他包中只需使用zap.L()调用即可
	zap.ReplaceGlobals(log0)
}
