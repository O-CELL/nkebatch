## Flasho sample tests
- [Test1](#test1)
- [Test2](#test2)
- [Test3](#test3)
- [Test4](#test4)
- [Test5](#test5)

# Test1 <a name="test1"></a>

- **Use case**:
    Mono serie, Separated Time stamp, Type float
- **Command**:
echo '10270080039320180080108183070d45851005' | ./br_uncompress -a  3 2,1.0,12

- **Buffer**:
    "buffer": [16,39,0,128,3,147,32,24,0,128,16,129,131,7,13,69,133,16,5],
- **Parameters**:
        "labelsize": 3,
        "series":
	    [{
	    "tag": 2,
	    "precision": 1.0,
	    "type": 12
        }]
 
- **Result**
<span style="color:grey"> 
    - UNCOMPRESS SERIE
    - cnt: 7
    - 1944
    - 1830 2 11.000000
    - 1845 2 13.000000
    - 1860 2 14.000000
    - 1875 2 21.000000
    - 1876 2 100.000000
</span>

## Test2 <a name="test2"></a>

- **Use case**:
    Two series, Separated Time stamp, Type uint32 and uint16
- **Command**:
echo « 26150020e06001d71e0000a0650f » | ./br_uncompress -a 1 0,1,10 1,100,6

- **Buffer**:
   - "buffer": [38,21,0,32,224,96,1,215,30,0,0,160,101,15],
- **Parameters**:
    "labelsize": 1,
    [{
    "tag": 0,
	"resolution": 1,
	"type": 10
    },{
	"tag": 1,
	"resolution": 100,
	"type": 6
    }]
 
- **Result**
<span style="color:grey"> 
    - UNCOMPRESS SERIE
    - cnt: 5
    - 263
    - 263 0 45
    - 263 1 3000
</span>

## Test3 <a name="test3"></a>


## Test4 <a name="test4"></a>
- **Use case**:
    Mono serie, Separated Time stamp, Type uint32, same value
- **Command**:
echo '1020c0220104a0214b351cb45b163b8965b72c$76cb62b72c76cb62b72cf6' | ./br_uncompress -a 1 0,1,10

- **Buffer**:
 "buffer": [16,32,192,34, 1, 4,160,33,75,53,28,180,91,22,59,137,101,183,44,118,203,98,183,44,118,203,98,183,44,246],
    
- **Parameters**:
    "labelsize": 1,
    "series":
	[{
	"tag": 0,
	"resolution": 1,
	"type": 10
    }]
   
 - **Result**
<span style="color:grey"> 
    - UNCOMPRESS SERIE
    - cnt: 0
    - 18226144
    - 18221344 0 874922
    - 18221944 0 874922
    - 18223144 0 874922
    - 18223744 0 874922
    - 18224344 0 874922
    - 18224944 0 874922
    - 18225544 0 874922
    - 18226144 0 874922

</span>

## Test5 <a name="test5"></a>

- **Use case**:
    Mono serie, Separated Time stamp, Type uint32, Different values
- **Command**:
echo '1027e0620a13a021ebb514b45bd667b72ca0ddb26ebb650577cbeaae' | ./br_uncompress -a 1 0,1,10

- **Buffer**:
    "buffer": [16,39,224,98,10,19,160,33,235,181,20,180,91,214,103,183,44,160,221,178,110,187,101, 5,119,203,234,174],
- **Parameters**:
    "labelsize": 1,
    "series":
	[{
	"tag": 0,
	"resolution": 1,
	"type": 10
    }]

 - **Result**
<span style="color:grey"> 
    - UNCOMPRESS SERIE
    - cnt: 7
    - 18308946
    - 18305944 0 874927
    - 18306544 0 874945
    - 18307144 0 875024
    - 18307744 0 875045
    - 18308344 0 875092
    - 18308944 0 875146
</span>
