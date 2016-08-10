##  Fuzz XML
This is a script built in go to fuzz any xml library concurrently 

Since you have two type of fuzzers available for you, you can choose between xmlfuzzer & radamsa :

## Install
```
Install xmlfuzzer:
http://komar.bitcheese.net/en/code/xmlfuzzer

Install radamsa:
git clone https://github.com/aoh/radamsa.git && cd radamsa && make && sudo make install

```
Then:
```
go get github.com/fatih/color
go build
```
Or you can use the compiled version:
```
./fuzzxml
```


## Usage
```
./fuzzxml [option] binary_to_be_fuzzed

Flags:
  -v	Show Version.
  -ra Use Radamsa fuzzer tool.
  -xf Use xmlfuzzer tool.
```


## Sample Output
![alt tag](https://git.corp.yahoo.com/raw/alaamub/fuzzxml/master/sample.png?token=AAAE5rYYZ8dNJUg7pll_xVXX9OcW2Avyks5XkVXLwA%3D%3D)
