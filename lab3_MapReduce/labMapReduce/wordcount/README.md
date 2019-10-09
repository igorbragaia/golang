./wordcount --mode sequential --file files/teste.txt --chunksize 100 --reducejobs 2

./wordcount --mode distributed --type worker --port 50001 --fail 3
./wordcount --mode distributed --type worker --port 50002

./wordcount --mode distributed --type master --file files/pg1342.txt --chunksize 102400 --reducejobs 5