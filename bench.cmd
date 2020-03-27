REM This command file runs the benchmarks presented in the README.md tables.
REM The benchmarks are in the package prng_test file prng_test.go.

SETLOCAL
REM cooling is the time in seconds to wait between each benchmark.
SET cooling=240

REM rounds is the number of iterations the benchmark loop will be executed.
SET  rounds=25000000000x
SET rounds2=15000000000x
SET rounds3=10000000000x

SET file=result.txt

ECHO Benchmarks started: %date%: %time% > %file%
TIMEOUT 1800
ECHO -------------------------------------- >> %file%
go test -run=zz prng_test -bench ^(BenchmarkEmpty)$ -benchtime=%rounds% >> %file%

ECHO -------------------------------------- >> %file%
go test -run=zz prng_test -bench ^(BenchmarkBaseline128)$ -benchtime=%rounds% >> %file%
TIMEOUT %cooling%

ECHO -------------------------------------- >> %file%
go test -run=zz prng_test -bench ^(BenchmarkState128)$ -benchtime=%rounds% >> %file%
TIMEOUT %cooling%

ECHO -------------------------------------- >> %file%
go test -run=zz prng_test -bench ^(BenchmarkBaseline256)$ -benchtime=%rounds% >> %file%
TIMEOUT %cooling%

ECHO -------------------------------------- >> %file%
go test -run=zz prng_test -bench ^(BenchmarkState256)$ -benchtime=%rounds% >> %file%
TIMEOUT %cooling%

ECHO -------------------------------------- >> %file%
go test -run=zz prng_test -bench ^(Benchmark128plus)$ -benchtime=%rounds% >> %file%
TIMEOUT %cooling%

ECHO -------------------------------------- >> %file%
go test -run=zz prng_test -bench ^(Benchmark128starstar)$ -benchtime=%rounds% >> %file%
TIMEOUT %cooling%

ECHO -------------------------------------- >> %file%
go test -run=zz prng_test -bench ^(BenchmarkSplitmix)$ -benchtime=%rounds% >> %file%
TIMEOUT %cooling%

ECHO -------------------------------------- >> %file%
go test -run=zz prng_test -bench ^(Benchmark256plus)$ -benchtime=%rounds% >> %file%
TIMEOUT %cooling%

ECHO -------------------------------------- >> %file%
go test -run=zz prng_test -bench ^(Benchmark256starstar)$ -benchtime=%rounds% >> %file%
TIMEOUT %cooling%

ECHO -------------------------------------- >> %file%
go test -run=zz prng_test -bench ^(Benchmark256plusplus)$ -benchtime=%rounds% >> %file%
TIMEOUT %cooling%

ECHO -------------------------------------- >> %file%
go test -run=zz prng_test -bench ^(Benchmark128sourceInterface)$ -benchtime=%rounds2% >> %file%
TIMEOUT %cooling%

ECHO -------------------------------------- >> %file%
go test -run=zz prng_test -bench ^(BenchmarkPCG)$ -benchtime=%rounds2% >> %file%
TIMEOUT %cooling%

ECHO -------------------------------------- >> %file%
go test -run=zz prng_test -bench ^(Benchmark256plusGonum)$ -benchtime=%rounds2% >> %file%
TIMEOUT %cooling%

ECHO -------------------------------------- >> %file%
go test -run=zz prng_test -bench ^(Benchmark256starstarGonum)$ -benchtime=%rounds2% >> %file%
TIMEOUT %cooling%

ECHO -------------------------------------- >> %file%
go test -run=zz prng_test -bench ^(BenchmarkInt63Rand)$ -benchtime=%rounds3% >> %file%
TIMEOUT %cooling%

ECHO -------------------------------------- >> %file%
go test -run=zz prng_test -bench ^(BenchmarkPCGsourceInterface)$ -benchtime=%rounds3% >> %file%
TIMEOUT %cooling%

ECHO -------------------------------------- >> %file%
go test -run=zz prng_test -bench ^(Benchmark256plusVpxyz)$ -benchtime=%rounds2% >> %file%
TIMEOUT %cooling%

ECHO -------------------------------------- >> %file%
go test -run=zz prng_test -bench ^(BenchmarkFloat64Xoro)$ -benchtime=%rounds2% >> %file%
TIMEOUT %cooling%

ECHO -------------------------------------- >> %file%
go test -run=zz prng_test -bench ^(BenchmarkFloat64_64)$ -benchtime=%rounds2% >> %file%
TIMEOUT %cooling%

ECHO -------------------------------------- >> %file%
go test -run=zz prng_test -bench ^(BenchmarkFloat64Xosh)$ -benchtime=%rounds2% >> %file%
TIMEOUT %cooling%

ECHO -------------------------------------- >> %file%
go test -run=zz prng_test -bench ^(BenchmarkFloat64Rand)$ -benchtime=%rounds3% >> %file%
TIMEOUT %cooling%

ECHO -------------------------------------- >> %file%
go test -run=zz prng_test -bench ^(BenchmarkFloat64NewRand)$ -benchtime=5000000000x >> %file%
TIMEOUT %cooling%

ECHO -------------------------------------- >> %file%
go test -run=zz prng_test -bench ^(BenchmarkPCGMTJ)$ -benchtime=%rounds3% >> %file%