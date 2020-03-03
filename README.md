# PrePregnancyAnno
annotation PrePregnancy

# get source
```
go get -d github.com/liserjrqlxue/PrePregnancyAnno
cd $GOPATH/src/github.com/liserjrqlxue/PrePregnancyAnno
```
or
```
# after get source.tar
tar avxf source.tar
cd PrePregnancyAnno
```
# generate db and codeKey
* first set private key code1 and code2 to protect your database
```
code1=118b09d39a5d3ecd56f9bd4f351dd6d6
code2=0e0760259f0826d18eb6e22988804617
```
* use util to aes encode your db
```
cd util
go build

mkdir -p ../db
# output to ../db/db.json.aes
./util -codeKey $code2 -excel /path/to/db.xlsx -prefix ../db/db

cd ..
```
* generate codeKey for users
```
cd generateKey
go build

./generateKey -code1 $code1 -code2 $code2 -user USERNAME
# save output codeKey to the USERNAME:
# Usr     DESKTOP-1S2D31U\A
# Code1   118b09d39a5d3ecd56f9bd4f351dd6d6
# Code2   0e0760259f0826d18eb6e22988804617
# Code3   1c4cc3ed9f4952f627d9f359319ab94c
# codeKey 610805cd69cfd3aa9ea23613b524c5a8807398172ef7011eac998442c62fb878

cd ..
```
* use the codeKey of your username to filter your anno and add extra db column 
```
go build

codeKey=610805cd69cfd3aa9ea23613b524c5a8807398172ef7011eac998442c62fb878
db=db/db.json.aes
keepDb=P100+F8
# output to /path/to/output.xlsx and /path/to/output.tsv
./PrePregnancyAnno -code $codeKey -aes $db -database $keepDb -var /path/to/input.anno -prefix /path/to/output
```