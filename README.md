# PrePregnancyAnno
annotation PrePregnancy

# get source
```
go get -d github.com/liserjrqlxue/PrePregnancyAnno
cd $GOPATH/src/github.com/liserjrqlxue/PrePregnancyAnno
```
or
```
// after get source.tar
tar avxf source.tar
cd PrePregnancyAnno
```
# generate db and codeKey
code1=118b09d39a5d3ecd56f9bd4f351dd6d6
code2=0e0760259f0826d18eb6e22988804617
cd util
go build
// output ../db/db.
./util -codeKey $code2 -excel /path/to/db.xlsx -prefix ../db/db
