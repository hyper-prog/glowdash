# GlowDash Script Language

GlowDash scripts are used in Action panels, worker functions in switches and the CommandLibrary functions.

> The scripts is line-by-line interpreted, every line can contains on one command without any terminator.
> Lines starting with `//` are comments.
> All variables are strings, but commands and operators interpret them as decimals or booleans as needed.

## Command Overview

| Command | Brief Description |
|---------|------------------|
| [If](#if) | Conditional execution |
| [Else](#else) | Alternate branch for If |
| [EndIf](#endif) | End of If block |
| [While](#while) | Loop while condition is true |
| [EndWhile](#endwhile) | End of While loop |
| [Return](#return) | Exit script (or function), optionally with a value |
| [PrintConsole](#printconsole) | Print text to the console |
| [Set](#set) | Set a variable to the value (And define if necessary) |
| [Run](#run) | Run a named ProgramLibrary element |
| [RunSet](#runset) | Run ProgramLibrary and store return value in a variable |
| [AddTo](#addto) | Add a value to a variable |
| [SubFrom](#subfrom) | Subtract a value from a variable |
| [MulWith](#mulwith) | Multiply a variable by a value |
| [DivWith](#divwith) | Divide a variable by a value |
| [RoundDown](#rounddown) | Round variable down |
| [RoundUp](#roundup) | Round variable up |
| [RoundMath](#roundmath) | Round variable mathematically |
| [AddMinutesToTime](#addminutestotime) | Add minutes to a time string |
| [WaitMs](#waitms) | Wait for milliseconds |
| [CallHttp](#callhttp) | Call an HTTP request (ignore result) |
| [CallHttpStoreJson](#callhttpstorejson) | Call HTTP and store JSON result |
| [SetFromJsonReq](#setfromjsonreq) | Extract a JSON element from HTTP response |
| [SetFromStoredJson](#setfromstoredjson) | Extract an element from stored JSON |
| [RelatedPanel](#relatedpanel) | Refresh related panels |
| [LoadVariablesFromPanelId](#loadvariablesfrompanelid) | Load variables from a panel |
| [LoadVariablesFromPanelIdWithPrefix](#loadvariablesfrompanelidwithprefix) | Load variables with a prefix |
| [PrintVariablesConsole](#printvariablesconsole) | Print all variables to the console |
| [AddOneshotSchedule](#addoneshotschedule) | Add a one-shot schedule |

## Operators for Expressions

| Operator | Meaning |
|----------|--------|
| ==       | Numerically equal |
| !=       | Numerically not equal |
| <        | Numerically less |
| <=       | Numerically less or equal |
| >        | Numerically greater |
| >=       | Numerically greater or equal |
| eq       | Strings are equal |
| neq      | Strings are not equal |
| in       | String is in a comma-separated list |
| nin      | String is not in a comma-separated list |
| booleq   | Compare as boolean |

> **Note:** Expressions in `If` and `While` only support two operands and a single operator. Parentheses and logical chaining (e.g., `and`, `or`) are not supported.

> **Note:** In all commands where an argument is requested, you can use variable substitution with `{{variablename}}`. For example, `If {{color}} == black`.

## Command Details (Unified)

### If
- **Syntax:** `If <expression>`
- **Parameters:**
  - `<expression>`: Logical condition using variables and operators. Only two operands and one operator are supported. Brackets and logical chaining are not allowed.
- **Description:** Starts a conditional block. If the expression is true, executes the following lines until `Else` or `EndIf`.
- **Sample:**
```glowdash
If {{state}} == open
    PrintConsole The state is open
Else
    PrintConsole The state is not open
EndIf
```

### While
- **Syntax:** `While <expression>`
- **Parameters:**
  - `<expression>`: Logical condition using variables and operators. Only two operands and one operator are supported. Brackets and logical chaining are not allowed.
- **Description:** Starts a loop. Executes the following lines while the expression is true, until `EndWhile`.
- **Sample:**
```glowdash
Set i 0
While {{i}} < 8
    PrintConsole In cycle, i is {{i}}
    AddTo i 1
EndWhile
```

### Return
- **Syntax:** `Return [value]`
- **Parameters:**
  - `[value]` (optional): Value to return from the script.
- **Description:** Exits the script, optionally returning a value.
- **Sample:**
```glowdash
Return

Return true

Set state 1
Return {{state}}
```

### PrintConsole
- **Syntax:** `PrintConsole <text>`
- **Parameters:**
  - `<text>`: Text or variable to print.
- **Description:** Prints text to the console/log.
- **Sample:**
```glowdash
PrintConsole Hello World

PrintConsole Variable name is {{name}}
```

### Set
- **Syntax:** `Set <variable> <value>`
- **Parameters:**
  - `<variable>`: Variable name. Supports `{{variablename}}` substitution.
  - `<value>`: Value to assign. Supports `{{variablename}}` substitution.
- **Description:** Sets a variable to a value.
- **Sample:**
```glowdash
Set catname Mirmur

Set catvar {{catname}}
```

### Run
- **Syntax:** `Run <programname>`
- **Parameters:**
  - `<programname>`: Name of ProgramLibrary element to run.
- **Description:** Runs a named ProgramLibrary element.
- **Sample:**
```glowdash
Run helloworld
```

### RunSet
- **Syntax:** `RunSet <variable> <programname>`
- **Parameters:**
  - `<variable>`: Variable to store return value.
  - `<programname>`: Name of ProgramLibrary element to run.
- **Description:** Runs a ProgramLibrary element and stores its return value.
- **Sample:**
```glowdash
RunSet result helloworld
```

### AddTo
- **Syntax:** `AddTo <variable> <value>`
- **Parameters:**
  - `<variable>`: Variable name.
  - `<value>`: Value to add.
- **Description:** Adds value to variable (numeric).
- **Sample:**
```glowdash
Set myvar 2
AddTo myvar 5
AddTo myvar {{increment}}
PrintConsole Variable myvar is {{myvar}}
```

### SubFrom
- **Syntax:** `SubFrom <variable> <value>`
- **Parameters:**
  - `<variable>`: Variable name.
  - `<value>`: Value to subtract.
- **Description:** Subtracts value from variable (numeric).
- **Sample:**
```glowdash
Set myvar 24
SubFrom myvar 2
```

### MulWith
- **Syntax:** `MulWith <variable> <value>`
- **Parameters:**
  - `<variable>`: Variable name.
  - `<value>`: Value to multiply.
- **Description:** Multiplies variable by value (numeric).
- **Sample:**
```glowdash
Set myvar 2
MulWith myvar 3
```

### DivWith
- **Syntax:** `DivWith <variable> <value>`
- **Parameters:**
  - `<variable>`: Variable name.
  - `<value>`: Value to divide.
- **Description:** Divides variable by value (numeric).
- **Sample:**
```glowdash
Set myvar 22
DivWith myvar 2
```

### AddMinutesToTime
- **Syntax:** `AddMinutesToTime <variable> <minutes>`
- **Parameters:**
  - `<variable>`: Variable containing time string (HH:MM).
  - `<minutes>`: Minutes to add.
- **Description:** Adds minutes to a time string.
- **Sample:**
```glowdash
Set time 14:45
AddMinutesToTime time 35
PrintConsole {{time}}
```

### WaitMs
- **Syntax:** `WaitMs <milliseconds>`
- **Parameters:**
  - `<milliseconds>`: Milliseconds to wait.
- **Description:** Pauses script for specified time.
- **Sample:**
```glowdash
WaitMs 1000

Set waitsec 5
MulWith waitsec 1000
WaitMs {{waitsec}}
```

### CallHttp
- **Syntax:** `CallHttp <url>`
- **Parameters:**
  - `<url>`: HTTP request URL.
- **Description:** Calls an HTTP request and ignores the result. After execution, the variable `LastHttpCallSuccess` is set to `true` if the request succeeded, or `false` if it failed.
- **Sample:**
```glowdash
CallHttp http://example.com/api

Set SwDeviceIp 192.168.1.22
CallHttp http://{{SwDeviceIp}}/rpc/Switch.Set?id=0&on=true
If {{LastHttpCallSuccess}} booleq false
    Return error
EndIf
```

### CallHttpStoreJson
- **Syntax:** `CallHttpStoreJson <variable> <url>`
- **Parameters:**
  - `<variable>`: Variable to store JSON result.
  - `<url>`: HTTP request URL.
- **Description:** Calls an HTTP request and stores the JSON result in a variable. This variables are different namespaces than other,
variables. They can not used in substitutions. After execution, the variable `LastHttpCallSuccess` is set to `true` if the request succeeded, or `false` if it failed.
- **Sample:**
```glowdash
CallHttpStoreJson myjson http://example.com/api

Set SwDeviceIp 192.168.1.22
CallHttpStoreJson myjson http://{{Sw.DeviceIp}}/rpc/Switch.GetStatus?id=0
If {{LastHttpCallSuccess}} booleq false
    Return error
EndIf

```

### SetFromJsonReq
- **Syntax:** `SetFromJsonReq <variable> <url> <jsonpath>`
- **Parameters:**
  - `<variable>`: Variable to store extracted value.
  - `<url>`: HTTP request URL.
  - `<jsonpath>`: JSON path to extract.
- **Description:** Calls an HTTP request, extracts the specified element from the JSON response, and stores it in a variable. After execution, the variable `LastHttpCallSuccess` is set to `true` if the request succeeded, or `false` if it failed.
- **Sample:**
```glowdash
SetFromJsonReq myvar http://example.com/api /element/path

Set SwDeviceIp 192.168.1.22
SetFromJsonReq myvar http://{{Sw.DeviceIp}}/rpc/Switch.GetStatus?id=0 /output
If {{LastHttpCallSuccess}} booleq false
    Return error
EndIf
Retrun {{myvar}}
```

### SetFromStoredJson
- **Syntax:** `SetFromStoredJson <variable> <jsonvar> <jsonpath>`
- **Parameters:**
  - `<variable>`: Variable to store extracted value.
  - `<jsonvar>`: Variable containing stored JSON. This variables are different namespaces than other, variables. They can not used in substitutions.
  - `<jsonpath>`: JSON path to extract.
- **Description:** Extracts element from previously stored JSON.
- **Sample:**
```glowdash
SetFromStoredJson myvar myjson /element/path

Set SwDeviceIp 192.168.1.22
CallHttpStoreJson myjson http://{{Sw.DeviceIp}}/rpc/Switch.GetStatus?id=0
If {{LastHttpCallSuccess}} booleq false
    Return error
EndIf
SetFromStoredJson myvar myjson /output
Retrun {{myvar}}
```

### RelatedPanel
- **Syntax:** `RelatedPanel <type> <ip> <deviceid>`
- **Parameters:**
  - `<type>`: Panel type.
  - `<ip>`: Device IP address.
  - `<deviceid>`: Device ID.
- **Description:** Refreshes related panels after script runs.
- **Sample:**
```glowdash
RelatedPanel Switch 192.168.1.101 0
```

### LoadVariablesFromPanelId
- **Syntax:** `LoadVariablesFromPanelId <panelid>`
- **Parameters:**
  - `<panelid>`: Panel ID to load variables from.
- **Description:** Loads exposed variables from the specified panel. For a Switch panel, the following variables may be loaded:

| Variable                   | Example Value |
|----------------------------|--------------|
| Panel.Id                   | sw1          |
| Panel.Title                | Living Room  |
| Panel.DeviceType           | Shelly       |
| Panel.SubPage              |              |
| Panel.Index                | 1            |
| Panel.PowerInfo            | true         |
| Panel.DeviceIp             | 192.168.1.10 |
| Panel.InDeviceId           | 0            |
| Panel.State                | 1            |
| Panel.InputState           | 1            |
| Panel.Watt                 | 12.50        |
| Panel.Volt                 | 230.00       |
| Panel.TextualState         | true         |
| Panel.TextualOppositeState | false        |

- **Sample:**
```glowdash
LoadVariablesFromPanelId sw1
// After execution, variables like Panel.Id, Panel.Title, Panel.DeviceType, etc. are available.
```

### LoadVariablesFromPanelIdWithPrefix
- **Syntax:** `LoadVariablesFromPanelIdWithPrefix <prefix> <panelid>`
- **Parameters:**
  - `<prefix>`: Prefix for loaded variables.
  - `<panelid>`: Panel ID to load variables from.
- **Description:** Loads exposed variables from the specified panel, adding the given prefix to each variable name. For a Switch panel with prefix `MySw_`, the following variables may be loaded:

| Variable                        | Example Value |
|----------------------------------|--------------|
| MySw_Panel.Id                   | sw1          |
| MySw_Panel.Title                | Living Room  |
| MySw_Panel.DeviceType           | Shelly       |
| MySw_Panel.SubPage              |              |
| MySw_Panel.Index                | 1            |
| MySw_Panel.PowerInfo            | true         |
| MySw_Panel.DeviceIp             | 192.168.1.10 |
| MySw_Panel.InDeviceId           | 0            |
| MySw_Panel.State                | 1            |
| MySw_Panel.InputState           | 1            |
| MySw_Panel.Watt                 | 12.50        |
| MySw_Panel.Volt                 | 230.00       |
| MySw_Panel.TextualState         | true         |
| MySw_Panel.TextualOppositeState | false        |

- **Sample:**
```glowdash
LoadVariablesFromPanelIdWithPrefix MySw_ sw1
// After execution, variables like MySw_Panel.Id, MySw_Panel.Title, etc. are available.
```

### PrintVariablesConsole
- **Syntax:** `PrintVariablesConsole`
- **Parameters:** None
- **Description:** Prints all variables to console/log.
- **Sample:**
```glowdash
PrintVariablesConsole
```

### RoundDown
- **Syntax:** `RoundDown <variable>`
- **Parameters:**
  - `<variable>`: Variable name containing a numeric value.
- **Description:** Rounds the variable down to the nearest integer.
- **Sample:**
```glowdash
Set myvar 5.12
RoundDown myvar
```

### RoundUp
- **Syntax:** `RoundUp <variable>`
- **Parameters:**
  - `<variable>`: Variable name containing a numeric value.
- **Description:** Rounds the variable up to the nearest integer.
- **Sample:**
```glowdash
Set myvar 5.12
RoundUp myvar
```

### RoundMath
- **Syntax:** `RoundMath <variable>`
- **Parameters:**
  - `<variable>`: Variable name containing a numeric value.
- **Description:** Rounds the variable to the nearest integer using standard mathematical rounding.
- **Sample:**
```glowdash
Set myvar 5.12
RoundMath myvar
```

### AddOneshotSchedule
- **Syntax:** `AddOneshotSchedule <panelid> <state> <time>`
- **Parameters:**
  - `<panelid>`: Panel ID to schedule.
  - `<state>`: State to set (e.g., 'on', 'off').
  - `<time>`: Time string (HH:MM) for the schedule.
- **Description:** Adds a one-shot schedule for the specified panel and state at the given time.
- **Sample:**
```glowdash
AddOneshotSchedule tswid001 off 15:30

Set swofftime {{Time.TimeHM}}
AddMinutesToTime swofftime 30
AddOneshotSchedule tswid001 off {{swofftime}}
```

## Predefined Variables

When a script starts, several variables related to the current date and time are automatically set and available for use. These are:

| Variable            | Description                                 | Example Value |
|---------------------|---------------------------------------------|--------------|
| Time.Hour           | Current hour (00-23)                        | 14           |
| Time.Minute         | Current minute (00-59)                      | 05           |
| Time.TimeHM         | Current time, hour and minute (HH:MM)        | 14:05        |
| Time.TimeHMS        | Current time, hour, minute, second (HH:MM:SS)| 14:05:23     |
| Time.Second         | Current second (00-59)                       | 23           |
| Time.SecOfDay       | Seconds since midnight                       | 50723        |
| Time.WeekDay        | Day of week (0=Sunday, 6=Saturday)           | 2            |
| Time.Month          | Current month (1-12)                         | 2            |
| Time.Day            | Current day of month (1-31)                  | 27           |
| Time.Year           | Current year                                 | 2026         |
| Time.YearDay        | Day of year (1-366)                          | 58           |

These variables are available in every script and can be used directly in expressions and commands.

---

### Sample Scripts
See [big-sample-config.yml](../config/big-sample-config.yml) for real-world examples.

```glowdash
RelatedPanel Shading 192.168.1.102 0
CallHttp http://192.168.1.102/rpc/Cover.Open?id=0
SetFromJsonReq cstate http://192.168.1.102/rpc/Cover.GetStatus?id=0 /state
While {{cstate}} nin open,stopped,closed
    WaitMs 1000
    SetFromJsonReq cstate http://192.168.1.102/rpc/Cover.GetStatus?id=0 /state
EndWhile
```

```glowdash
RelatedPanel Shading {{Panel.DeviceIp}} {{Panel.InDeviceId}}
Set BackOpenTime 300
Set DeviceIp {{Panel.DeviceIp}}
Set DeviceId {{Panel.InDeviceId}}
Set MinDownPos 70
SetFromJsonReq CoverPos http://{{DeviceIp}}/rpc/Cover.GetStatus?id={{DeviceId}} /current_pos
//PrintConsole CoverPos is {{CoverPos}} % ({{TIME.HOUR}}:{{TIME.MINUTE}})
Set MinDownPosPlus2 {{MinDownPos}}
AddTo MinDownPosPlus2 2
If {{CoverPos}} > {{MinDownPosPlus2}}
    //Smaller than min%, down to min%.
    CallHttp http://{{DeviceIp}}/rpc/Cover.GoToPosition?id={{DeviceId}}&pos={{MinDownPos}}
    Set cstate none
    While {{cstate}} nin open,stopped,closed
        WaitMs 1000
        SetFromJsonReq cstate http://{{DeviceIp}}/rpc/Cover.GetStatus?id={{DeviceId}} /state
    EndWhile
Else
    CallHttp http://{{DeviceIp}}/rpc/Cover.Close?id={{DeviceId}}
    WaitMs 2000
    CallHttp http://{{DeviceIp}}/rpc/Cover.Stop?id={{DeviceId}}
EndIf
WaitMs 200
CallHttp http://{{DeviceIp}}/rpc/Cover.Open?id={{DeviceId}}
WaitMs {{BackOpenTime}}
CallHttp http://{{DeviceIp}}/rpc/Cover.Stop?id={{DeviceId}}
```

---
