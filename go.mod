module github.com/pzqf/zGameServer

go 1.25.5

require (
	github.com/pzqf/zEngine v0.0.1
	github.com/pzqf/zUtil v0.0.1
	go.uber.org/zap v1.21.0
	google.golang.org/protobuf v1.26.0
	gopkg.in/ini.v1 v1.67.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/go-sql-driver/mysql v1.9.3 // indirect
	github.com/panjf2000/ants v1.3.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
)

replace github.com/pzqf/zEngine => ../zEngine

replace github.com/pzqf/zUtil => ../zUtil
