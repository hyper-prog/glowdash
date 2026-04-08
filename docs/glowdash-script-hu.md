# GlowDash szkriptnyelv

A GlowDash szkriptek az Action panelekben, a kapcsolók worker függvényeiben és a CommandLibrary függvényeiben használhatók.

> A szkriptek soronként kerülnek értelmezésre, minden sor legfeljebb egy parancsot tartalmazhat lezáró karakter nélkül.
> A `//` jellel kezdődő sorok megjegyzések.
> Minden változó string, de a parancsok és operátorok szükség szerint decimálisként vagy logikai értékként értelmezik őket.

## Parancsok áttekintése

| Command | Brief Description |
|---------|------------------|
| [If](#if) | Feltételes végrehajtás |
| [Else](#else) | Alternatív ág az If-hez |
| [EndIf](#endif) | Az If blokk vége |
| [While](#while) | Ciklus, amíg a feltétel igaz |
| [EndWhile](#endwhile) | A While ciklus vége |
| [Return](#return) | Kilépés a szkriptből (vagy függvényből), opcionálisan értékkel |
| [PrintConsole](#printconsole) | Szöveg kiírása a konzolra (standard output) |
| [PrintGlowdashConsole](#printconsole) | Szöveg kiírása a GlowDash konzolra |
| [Set](#set) | Változó értékének beállítása (és szükség esetén létrehozása) |
| [Run](#run) | Név szerinti ProgramLibrary elem futtatása |
| [RunSet](#runset) | ProgramLibrary futtatása és visszatérési érték mentése változóba |
| [AddTo](#addto) | Érték hozzáadása változóhoz |
| [SubFrom](#subfrom) | Érték kivonása változóból |
| [MulWith](#mulwith) | Változó szorzása értékkel |
| [DivWith](#divwith) | Változó osztása értékkel |
| [RoundDown](#rounddown) | Változó lefelé kerekítése |
| [RoundUp](#roundup) | Változó felfelé kerekítése |
| [RoundMath](#roundmath) | Változó matematikai kerekítése |
| [AddMinutesToTime](#addminutestotime) | Perc hozzáadása idő karakterlánchoz |
| [WaitMs](#waitms) | Várakozás milliszekundumban |
| [CallHttp](#callhttp) | HTTP kérés hívása (eredmény figyelmen kívül hagyása) |
| [CallHttpStoreJson](#callhttpstorejson) | HTTP hívás és JSON eredmény eltárolása |
| [SetFromJsonReq](#setfromjsonreq) | JSON elem kinyerése HTTP válaszból |
| [SetFromStoredJson](#setfromstoredjson) | Elem kinyerése eltárolt JSON-ból |
| [RelatedPanel](#relatedpanel) | Kapcsolódó panelek frissítése |
| [LoadVariablesFromPanelId](#loadvariablesfrompanelid) | Változók betöltése panelből |
| [LoadVariablesFromPanelIdWithPrefix](#loadvariablesfrompanelidwithprefix) | Változók betöltése előtaggal |
| [PrintVariablesConsole](#printvariablesconsole) | Összes változó kiírása a konzolra (standard output) |
| [PrintVariablesGlowdashConsole](#printvariablesglowdashconsole) | Összes változó kiírása a GlowDash konzolra |
| [AddOneshotSchedule](#addoneshotschedule) | Egyszeri ütemezés hozzáadása |
| [ModbusTcp](#modbustcp) | Modbus TCP regiszter vagy coil olvasása/írása |
| [ShellyRelay](#shellyrelay) | Shelly eszközök olvasása/írása (relay/cover) |


## Futásidejű állapotváltozók (`state.` előtag)

A GlowDash szkriptek alapvetően állapotmentesek: minden szkriptfutás egymástól független,
és a lokális változók nem maradnak meg események között.
Néhány esetben azonban szükséges adatot megosztani különböző szkriptfutások között a program futása alatt
(például virtuális kapcsolók megvalósításához vagy köztes értékek gyorsítótárazásához).

Ennek támogatására a GlowDash futásidejű állapotváltozókat biztosít, amelyek a `state.` előtagon keresztül érhetők el.

### Áttekintés

Azok a változók, amelyek neve `state.` előtaggal kezdődik, egy megosztott, memóriabeli tárolóra mutatnak, amely:

- **Globális a futó GlowDash példányon belül**
- **Minden szkriptből elérhető**
- **A szkriptfutások között is megmarad**
- **Csak a program futása alatt létezik (lemezre nem mentődik)**
- **A GlowDash újraindításakor törlődik**

Ez lehetővé teszi, hogy különböző események által indított szkriptek biztonságosan adatot cseréljenek tartós tároló bevezetése nélkül.

### State változók olvasása és írása

A state változók ugyanúgy használhatók, mint a normál változók, csak `state.` előtaggal.

```glowdash
Set state.myflag true

If state.myflag booleq true
    ...
EndIf
```
### Kezdeti értékek beállítása `state.` változókhoz

A Glowdash indulásakor a `GlowdashStart` könyvtári függvény automatikusan elindul.
Ott beállíthatók a `state.` változók.

### Ajánlott felhasználási esetek

A futásidejű állapotváltozók hasznosak például:

- Virtuális eszközökhöz (pl. szoftveres kapcsolók)
- Eseménykoordinációhoz szkriptek között
- Ideiglenes jelzőkhöz vagy számlálókhoz
- Egyszerű memóriabeli gyorsítótárazáshoz
- Értékek megjegyzéséhez triggerek között

Nem tartós adattárolásra valók.
Tartós konfigurációhoz vagy hosszú távú adatokhoz külső tárolási mechanizmus használata javasolt.


## Operátorok kifejezésekhez

A kifejezések az `If` és `While` feltételeiben használhatók. Két forma létezik:
- **1 operandusos kifejezés:** `<operator> <operand>` — először az operátor, utána egyetlen érték vagy változó.
- **2 operandusos kifejezés:** `<operand1> <operator> <operand2>` — érték vagy változó mindkét oldalon.

Az egyértékes (operátor nélküli) kifejezések is elfogadottak: egy változó vagy literál önmagában logikai értékként kerül kiértékelésre.

> **Note:** A zárójelek és a logikai láncolás (pl. `and`, `or`) nem támogatott.

> **Note:** Minden parancsban, ahol argumentum szükséges, használható változóhelyettesítés `{{variablename}}` formában. Például: `If {{color}} eq black`.

### 1 operandusos operátorok

Szintaxis: `<operator> <value>`

| Operator     | Meaning |
|--------------|---------|
| not          | A `<value>` logikai értékének negálása |
| isEmpty      | Igaz, ha `<value>` üres vagy csak szóközöket tartalmaz |
| isNotEmpty   | Igaz, ha `<value>` nem üres |
| isDefined    | Igaz, ha a `<value>` nevű változó definiálva van |
| isNotDefined | Igaz, ha a `<value>` nevű változó nincs definiálva |

**Samples:**
```glowdash
If isEmpty {{myvar}}
    PrintGlowdashConsole myvar is empty
EndIf

If isNotEmpty {{myvar}}
    PrintGlowdashConsole myvar has a value
EndIf

If isDefined state.myflag
    PrintGlowdashConsole state.myflag is defined
EndIf

If isNotDefined myvar
    Set myvar defaultvalue
EndIf

If not {{myflag}}
    PrintConsole myflag is falsy
EndIf
```

### 2 operandusos operátorok

Szintaxis: `<value1> <operator> <value2>`

| Operator | Meaning |
|----------|---------|
| ==       | Numerikusan egyenlő |
| !=       | Numerikusan nem egyenlő |
| <        | Numerikusan kisebb |
| <=       | Numerikusan kisebb vagy egyenlő |
| >        | Numerikusan nagyobb |
| >=       | Numerikusan nagyobb vagy egyenlő |
| eq       | Stringek megegyeznek |
| neq      | Stringek nem egyeznek |
| in       | A string benne van egy vesszővel elválasztott listában |
| nin      | A string nincs benne egy vesszővel elválasztott listában |
| booleq   | Összehasonlítás logikai értékként |

## Parancsok részletesen (egységesen)

### If
- **Syntax:** `If <expression>`
- **Parameters:**
  - `<expression>`: Logikai feltétel. Támogatott formák: 1 operandusos (`<operator> <value>`), 2 operandusos (`<value> <operator> <value>`) vagy egyértékes logikai forma. Zárójelek és logikai láncolás nem engedélyezett.
- **Description:** Feltételes blokk indítása. Ha a kifejezés igaz, a következő sorokat végrehajtja `Else` vagy `EndIf` sorig.
- **Sample:**
```glowdash
If {{count}} == 2
    PrintConsole There is two
Else
    PrintConsole There is not two
EndIf

If {{color}} eq black
    PrintConsole The color is black
EndIf

If isEmpty {{myvar}}
    Set myvar default
EndIf

If isDefined state.myflag
    PrintConsole flag is set
EndIf
```

### While
- **Syntax:** `While <expression>`
- **Parameters:**
  - `<expression>`: Logikai feltétel. Támogatott formák: 1 operandusos (`<operator> <value>`), 2 operandusos (`<value> <operator> <value>`) vagy egyértékes logikai forma. Zárójelek és logikai láncolás nem engedélyezett.
- **Description:** Ciklus indítása. A következő sorok végrehajtása addig, amíg a kifejezés igaz, `EndWhile` sorig.
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
  - `[value]` (optional): A szkriptből visszaadandó érték.
- **Description:** Kilépés a szkriptből, opcionálisan érték visszaadásával.
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
  - `<text>`: Kiírandó szöveg vagy változó.
- **Description:** Szöveg kiírása a konzolba/naplóba. Standard output.
- **Sample:**
```glowdash
PrintConsole Hello World

PrintConsole Variable name is {{name}}
```

### PrintGlowdashConsole
- **Syntax:** `PrintGlowdashConsole <text>`
- **Parameters:**
  - `<text>`: Kiírandó szöveg vagy változó.
- **Description:** Szöveg kiírása a Glowdash konzolba.
- **Sample:**
```glowdash
PrintGlowdashConsole Hello World

PrintGlowdashConsole Variable name is {{name}}
```

### Set
- **Syntax:** `Set <variable> <value>`
- **Parameters:**
  - `<variable>`: Változónév. Támogatja a `{{variablename}}` helyettesítést.
  - `<value>`: Beállítandó érték. Támogatja a `{{variablename}}` helyettesítést.
- **Description:** Változó definiálása és értékadása. Már definiált változók értéke is módosítható.
- **Sample:**
```glowdash
Set catname Mirmur

Set catvar {{catname}}
Set catvar 3
```

### Run
- **Syntax:** `Run <programname>`
- **Parameters:**
  - `<programname>`: A futtatandó ProgramLibrary elem neve.
- **Description:** Név szerinti ProgramLibrary elem futtatása.
- **Sample:**
```glowdash
Run helloworld
```

### RunSet
- **Syntax:** `RunSet <variable> <programname>`
- **Parameters:**
  - `<variable>`: Változó a visszatérési érték tárolására.
  - `<programname>`: A futtatandó ProgramLibrary elem neve.
- **Description:** ProgramLibrary elem futtatása és a visszatérési érték eltárolása.
- **Sample:**
```glowdash
RunSet result helloworld
```

### AddTo
- **Syntax:** `AddTo <variable> <value>`
- **Parameters:**
  - `<variable>`: Változónév.
  - `<value>`: A hozzáadandó érték.
- **Description:** Érték hozzáadása változóhoz (numerikus).
- **Sample:**
```glowdash
Set myvar 2
// myvar is 2

AddTo myvar 5
// myvar is 5

Set increment 3
AddTo myvar {{increment}}
// myvar is 8

PrintConsole Variable myvar is {{myvar}}
```

### SubFrom
- **Syntax:** `SubFrom <variable> <value>`
- **Parameters:**
  - `<variable>`: Változónév.
  - `<value>`: A kivonandó érték.
- **Description:** Érték kivonása változóból (numerikus).
- **Sample:**
```glowdash
Set myvar 24
SubFrom myvar 2
// myvar is 22
```

### MulWith
- **Syntax:** `MulWith <variable> <value>`
- **Parameters:**
  - `<variable>`: Változónév.
  - `<value>`: A szorzó érték.
- **Description:** Változó szorzása értékkel (numerikus).
- **Sample:**
```glowdash
Set myvar 2
MulWith myvar 3
// myvar is 6
```

### DivWith
- **Syntax:** `DivWith <variable> <value>`
- **Parameters:**
  - `<variable>`: Változónév.
  - `<value>`: Az osztó érték.
- **Description:** Változó osztása értékkel (numerikus).
- **Sample:**
```glowdash
Set myvar 22
DivWith myvar 2
// myvar 11
```

### AddMinutesToTime
- **Syntax:** `AddMinutesToTime <variable> <minutes>`
- **Parameters:**
  - `<variable>`: Időt tartalmazó változó (HH:MM).
  - `<minutes>`: Hozzáadandó percek.
- **Description:** Percek hozzáadása idő karakterlánchoz.
- **Sample:**
```glowdash
Set time 14:45
AddMinutesToTime time 35
PrintConsole {{time}}
// time is 15:20
```

### WaitMs
- **Syntax:** `WaitMs <milliseconds>`
- **Parameters:**
  - `<milliseconds>`: Várakozási idő ezredmásodpercben.
- **Description:** A szkript szüneteltetése a megadott ideig.
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
  - `<url>`: HTTP kérés URL-je.
- **Description:** HTTP kérés végrehajtása, az eredmény figyelmen kívül hagyásával. Futás után a `LastHttpCallSuccess` változó `true`, ha a kérés sikerült, vagy `false`, ha sikertelen volt.
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
  - `<variable>`: Változó a JSON eredmény tárolására.
  - `<url>`: HTTP kérés URL-je.
- **Description:** HTTP kérés végrehajtása és a JSON eredmény tárolása változóban. Ezek a változók eltérő névtérben vannak a többi változóhoz képest,
változók. Nem használhatók helyettesítésben. Futás után a `LastHttpCallSuccess` változó `true`, ha a kérés sikerült, vagy `false`, ha sikertelen volt.
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
  - `<variable>`: Változó a kinyert érték tárolására.
  - `<url>`: HTTP kérés URL-je.
  - `<jsonpath>`: A kinyerendő JSON útvonal.
- **Description:** HTTP kérés végrehajtása, a megadott elem kinyerése a JSON válaszból, majd eltárolása változóban. Futás után a `LastHttpCallSuccess` változó `true`, ha a kérés sikerült, vagy `false`, ha sikertelen volt.
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
  - `<variable>`: Változó a kinyert érték tárolására.
  - `<jsonvar>`: Eltárolt JSON-t tartalmazó változó. Ezek a változók eltérő névtérben vannak a többi változóhoz képest. Nem használhatók helyettesítésben.
  - `<jsonpath>`: A kinyerendő JSON útvonal.
- **Description:** Elem kinyerése egy korábban eltárolt JSON-ból.
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
  - `<type>`: Panel típusa.
  - `<ip>`: Eszköz IP címe.
  - `<deviceid>`: Eszköz azonosítója.
- **Description:** Kapcsolódó panelek frissítése a szkript futása után.
- **Sample:**
```glowdash
RelatedPanel Switch 192.168.1.101 0
```

### LoadVariablesFromPanelId
- **Syntax:** `LoadVariablesFromPanelId <panelid>`
- **Parameters:**
  - `<panelid>`: A panel ID, amelyből a változókat be kell tölteni.
- **Description:** A megadott panelen publikált változók betöltése. Switch panel esetén például az alábbi változók tölthetők be:

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
  - `<prefix>`: Előtag a betöltött változókhoz.
  - `<panelid>`: A panel ID, amelyből a változókat be kell tölteni.
- **Description:** A megadott panelen publikált változók betöltése úgy, hogy minden változónévhez hozzáadja a megadott előtagot. Például `MySw_` előtagú Switch panel esetén az alábbi változók tölthetők be:

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
- **Description:** Az összes változó kiírása a konzolba/naplóba. Standard output.
- **Sample:**
```glowdash
PrintVariablesConsole
```

### PrintVariablesGlowdashConsole
- **Syntax:** `PrintVariablesGlowdashConsole`
- **Parameters:** None
- **Description:** Az összes változó kiírása a GlowDash konzolra.
- **Sample:**
```glowdash
PrintVariablesGlowdashConsole
```

### RoundDown
- **Syntax:** `RoundDown <variable>`
- **Parameters:**
  - `<variable>`: Numerikus értéket tartalmazó változó neve.
- **Description:** A változó lefelé kerekítése a legközelebbi egészre.
- **Sample:**
```glowdash
Set myvar 5.12
RoundDown myvar
// myvar is 5.0
```

### RoundUp
- **Syntax:** `RoundUp <variable>`
- **Parameters:**
  - `<variable>`: Numerikus értéket tartalmazó változó neve.
- **Description:** A változó felfelé kerekítése a legközelebbi egészre.
- **Sample:**
```glowdash
Set myvar 5.12
RoundUp myvar
// myvar is 6.0
```

### RoundMath
- **Syntax:** `RoundMath <variable>`
- **Parameters:**
  - `<variable>`: Numerikus értéket tartalmazó változó neve.
- **Description:** A változó kerekítése a legközelebbi egészre, szabványos matematikai kerekítéssel.
- **Sample:**
```glowdash
Set myvar 5.12
RoundMath myvar
// myvar is 5.0

Set myvar 3.72
RoundMath myvar
// myvar is 4.0
```

### AddOneshotSchedule
- **Syntax:** `AddOneshotSchedule <panelid> <action> <time>`
- **Parameters:**
  - `<panelid>`: Az ütemezendő panel azonosítója.
  - `<action>`: A panelen aktiválandó művelet (pl. kapcsolóknál 'on', 'off', műveleteknél 'run').
  - `<time>`: Idő karakterlánc (HH:MM) az ütemezéshez.
- **Description:** Egyszeri ütemezés hozzáadása a megadott panelhez és állapothoz a megadott időpontban.
- **Sample:**
```glowdash
AddOneshotSchedule tswid001 off 15:30

Set swofftime {{Time.TimeHM}}
AddMinutesToTime swofftime 30
AddOneshotSchedule tswid001 off {{swofftime}}

AddOneshotSchedule ac004 run 19:15
```

### ModbusTcp
- **Syntax:** `ModbusTcp <variable> <host:port> <unitId> <operation> <address>`
- **Parameters:**
  - `<variable>`: Olvasási műveletnél az eredmény tárolására szolgáló változónév. Írási műveletnél az írandó változó vagy literál érték.
  - `<host:port>`: A Modbus TCP eszköz IP címe és portja (pl. `192.168.1.50:502`).
  - `<unitId>`: Modbus unit ID (slave cím), tipikusan `1`.
  - `<operation>`: Egyik az alábbiak közül: `readcoil`, `readinput`, `readregister`, `writecoil`, `writeregister`.
  - `<address>`: Regiszter vagy coil cím (decimális).
- **Description:** Kommunikáció Modbus TCP eszközzel. Futás után a `LastModbusTcpCallSuccess` változó `true`, ha a művelet sikerült, vagy `false`, ha sikertelen volt.

| Operation       | Description                                              | Result stored in variable |
|-----------------|----------------------------------------------------------|---------------------------|
| `readcoil`      | Egyetlen coil olvasása (FC 0x01), eredmény `true`/`false`  | yes |
| `readinput`     | Egyetlen input regiszter olvasása (FC 0x04), eredmény decimális egész | yes |
| `writecoil`     | Egyetlen coil írása (FC 0x05), érték a `<variable>` alapján  | no |

- **Sample:**
```glowdash
// Read a coil at address 5 from device at 192.168.1.50, unit 1
ModbusTcp coilstate 192.168.1.50:502 1 readcoil 5
If not {{LastModbusTcpCallSuccess}}
    PrintGlowdashConsole Modbus read failed
EndIf
PrintGlowdashConsole Coil state: {{coilstate}}
Return {{coilstate}}

// Read input register at address 200
ModbusTcp sensorval 192.168.1.50:502 1 readinput 200
PrintConsole Sensor value: {{sensorval}}

// Write a coil at address 3 with value from variable
Set relaystate true
ModbusTcp {{relaystate}} 192.168.1.50:502 1 writecoil 3
```

### ShellyRelay
- **Syntax:**
  - `ShellyRelay <variable> <host[:port]> readrelay <inDeviceId>`
  - `ShellyRelay <variable> <host[:port]> readcover <inDeviceId>`
  - `ShellyRelay <value> <host[:port]> setrelay <inDeviceId>`
  - `ShellyRelay <action> <host[:port]> setcover <inDeviceId>`
- **Parameters:**
  - `<variable>`: Olvasási műveleteknél az eredmény tárolására szolgáló változónév.
  - `<value>`: `setrelay` esetén logikai jellegű érték (`true`, `false`, `1`, `0`, `on`, `off`, stb.).
  - `<action>`: `setcover` esetén a Shelly cover action kezelőnek átadott parancs string.
  - `<host[:port]>`: Eszköz IP/host, opcionális porttal. Ha a port nincs megadva, alapértelmezett érték `80`.
  - `<operation>`: Az alábbiak egyike: `readrelay`, `setrelay`, `readcover`, `setcover`.
  - `<inDeviceId>`: Shelly csatorna/index (például `0`, `1`, ...).
- **Description:** Kommunikáció Shelly relé vagy cover végponttal. Futás után a `LastShellyRelayCallSuccess` változó `true`, ha a művelet sikerült, vagy `false`, ha sikertelen volt.

| Operation     | Description | Variables set |
|---------------|-------------|---------------|
| `readrelay`   | Reléállapot és bemeneti állapot olvasása | `<variable>` (`true`/`false`), `<variable>.StateInt` (`0`/`1`), `<variable>.StateBool` (`true`/`false`), `<variable>.InputState` (`true`/`false`) |
| `setrelay`    | Relé be-/kikapcsolása `<value>` alapján | nincs műveletspecifikus kimeneti változó |
| `readcover`   | Cover pozíció és név szerinti állapot olvasása | `<variable>` (pozíció egész), `<variable>.Position`, `<variable>.NamedState` |
| `setcover`    | Cover action parancs küldése `<action>` alapján | nincs műveletspecifikus kimeneti változó |

- **Sample:**
```glowdash
// Read relay state from channel 0
ShellyRelay relayState 192.168.1.22 readrelay 0
If not {{LastShellyRelayCallSuccess}}
    PrintGlowdashConsole Shelly relay read failed
    Return error
EndIf
PrintGlowdashConsole Relay state: {{relayState}} (input={{relayState.InputState}})

// Set relay on
Set desiredState true
ShellyRelay {{desiredState}} 192.168.1.22 setrelay 0

// Read cover status
ShellyRelay coverPos 192.168.1.23 readcover 1
PrintConsole Cover position: {{coverPos}} ({{coverPos.NamedState}})

// Control cover
ShellyRelay open 192.168.1.23 setcover 1
```

## Előre definiált változók

Amikor egy szkript elindul, több, az aktuális dátumhoz és időhöz kapcsolódó változó automatikusan beállításra kerül és használhatóvá válik. Ezek a következők:

| Variable            | Description                                  | Example Value |
|---------------------|----------------------------------------------|--------------|
| Time.Hour           | Aktuális óra (00-23)                         | 14           |
| Time.Minute         | Aktuális perc (00-59)                        | 05           |
| Time.TimeHM         | Aktuális idő, óra és perc (HH:MM)            | 14:05        |
| Time.TimeHMS        | Aktuális idő, óra, perc, másodperc (HH:MM:SS)| 14:05:23     |
| Time.Second         | Aktuális másodperc (00-59)                   | 23           |
| Time.SecOfDay       | Másodpercek száma éjfél óta                  | 50723        |
| Time.WeekDay        | A hét napja (0=vasárnap, 6=szombat)          | 2            |
| Time.Month          | Aktuális hónap (1-12)                        | 2            |
| Time.Day            | Aktuális nap a hónapban (1-31)               | 27           |
| Time.Year           | Aktuális év                                  | 2026         |
| Time.YearDay        | Az év napja (1-366)                          | 58           |

Ezek a változók minden szkriptben elérhetők, és közvetlenül használhatók kifejezésekben és parancsokban.

---

### Mintaszkriptek
Valós példákért lásd: [big-sample-config.yml](../config/big-sample-config.yml).

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
