/* GlowDash - Smart Home Web Dashboard
   (C) 2024-2025 Péter Deák (hyper80@gmail.com)
   License: GPLv2
*/

const thermostatTempColors = [
    { temp: 5, color: '#4040ff' },
    { temp: 16, color: '#a080ff' },
    { temp: 18, color: '#ffff00' },
    { temp: 20, color: '#00ff00' },
    { temp: 23, color: '#e6b207' },
    { temp: 25, color: '#ff0000' },
    { temp: 99, color: '#ff0000' }
];

var xhrRequests = new Map();

function runCommand(cmd) {
    const parts = cmd.split(":");
    if(parts[0] == "sethtml" && parts.length == 3) {
        let content = b64DecodeUnicode(parts[2]);
        document.querySelectorAll(window.atob(parts[1])).forEach(item => {
            item.innerHTML = content;
        });
        initializeActions();
        handleJustMovingPanels();
        initUnravedGauges();
        initClockPickerBlocks();
        initActionSubselector();
        return;
    }
    if(parts[0] == "loadpage" && parts.length == 2) {
        window.location = window.atob(parts[1]);
        return;
    }
    if(parts[0] == "refreshpage" && parts.length == 1) {
        location.reload();
        return;
    }
}

function b64DecodeUnicode(str) {
    return decodeURIComponent(atob(str).split('').map(function(c) {
        return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
    }).join(''));
}

function abortXhrRequestsForGrpId(grpId) {
    if(xhrRequests.has(grpId + "-up")) {
        let xhrReq = xhrRequests.get(grpId + "-up");
        xhrReq.abort();
        xhrRequests.delete(grpId + "-up");
    }
    if(xhrRequests.has(grpId + "-down")) {
        let xhrReq = xhrRequests.get(grpId + "-down");
        xhrReq.abort();
        xhrRequests.delete(grpId + "-down");
    }
    if(xhrRequests.has(grpId + "-update")) {
        let xhrReq = xhrRequests.get(grpId + "-update");
        xhrReq.abort();
        xhrRequests.delete(grpId + "-update");
    }
}

var skip_therm_sse_requests = false;
var thermostatWaitTimeoutMillisec = 6 * 1000;
var thermoBtnSetWaitId = 0;

function thermostatSetTimeoutReach(id) {
    let val = document.getElementById("thermostatTemperatureDisplay-" + id).innerHTML;
    let xhr = new XMLHttpRequest();
    xhrRequests.set(id,xhr);

    xhr.open('GET', '/action/b-'+id+"-tts/"+val+"?otsseid="+ot_sse_id, true);
    skip_therm_sse_requests = true;
    xhr.onreadystatechange = function(e) {
        xhrRequests.delete(id);
        if (xhr.readyState == 4 && xhr.status == 200) {
            if(xhr.responseText == "")
                return;
            skip_therm_sse_requests = false;
            let js = JSON.parse(xhr.responseText);
            if(js["result"] == "ok") {
                for(i = 0; i < js["cmds"].length ; i++)
                    runCommand(js["cmds"][i]);
            }
        }
    };
    xhr.send();
    if(thermoBtnSetWaitId != 0)
        clearInterval(thermoBtnSetWaitId);
    thermoBtnSetWaitId = 0;
}

function initializeActions() {
    const allJsActionButton = document.getElementsByClassName("jsaction");
    for (let i = 0; i < allJsActionButton.length; i++) {
        let id = allJsActionButton[i].id;
        if(allJsActionButton[i].classList.contains('jsact-processed'))
            continue;

        allJsActionButton[i].addEventListener("click", function(e) {
            let targetElement = document.getElementById(id);
            let animatedElementId = id;

            //Special handling for thermostat buttons (On pane type "Thermostat" )
            if(targetElement.classList.contains('thermobtn')) {
                let grpId = targetElement.dataset.grpid;
                if(grpId.substr(0,2) != "b-")
                    return;
                let panelId = grpId.substr(2);
                let fvalStr = document.getElementById("thermostatTemperatureDisplay-" + panelId).dataset.fragval;
                let fvalInt = parseInt(fvalStr);
                if(grpId+"-up" == id) {
                    fvalInt += 1;
                }
                if(grpId+"-down" == id) {
                    fvalInt -= 1;
                }

                //Removing animation class and stop running animation
                document.querySelectorAll("#pc-"+panelId+" .sendthermsign.animasendtherm").forEach(item => {
                    item.classList.remove("animasendtherm");
                    item.getAnimations().forEach((anim) => {
                        anim.cancel();
                    });
                });

                setTemperature(5.0 + fvalInt*0.5,panelId);
                if(thermoBtnSetWaitId != 0)
                    clearInterval(thermoBtnSetWaitId);
                thermoBtnSetWaitId = setInterval(thermostatSetTimeoutReach,thermostatWaitTimeoutMillisec,panelId);

                //Add animation class
                document.querySelectorAll("#pc-"+panelId+" .sendthermsign").forEach(item => {
                    item.classList.add("animasendtherm");
                });
                return;
            }

            //Special handling for shading up or down buttons (On pane type "Shading" )
            if(targetElement.classList.contains('zcombomove')) {
                let stopButtonId = targetElement.dataset.grpid + "-stop";
                document.getElementById(id).classList.add("displaynone");
                document.getElementById(stopButtonId).classList.remove("displaynone");
                if(id.endsWith("-up"))
                    document.getElementById(targetElement.dataset.grpid + "-down").classList.remove("displaynone");
                if(id.endsWith("-down"))
                    document.getElementById(targetElement.dataset.grpid + "-up").classList.remove("displaynone");
                animatedElementId = stopButtonId;
                abortXhrRequestsForGrpId(targetElement.dataset.grpid)
            }

            //Special handling for shading stop buttons (On pane type "Shading" )
            if(targetElement.classList.contains('zcombostop')) {
                abortXhrRequestsForGrpId(targetElement.dataset.grpid)
            }

            //Special handling for toggle switch buttons (On pane type "ToggleSwitch" )
            if(targetElement.classList.contains('tglswbtn')) {
                let tgshId = targetElement.dataset.tgshid;
                if(tgshId.substr(0,9) != "tglshows-")
                    return;
                //Start flip animation
                document.getElementById(tgshId + "-tpp").classList.add("circle-avatar-flip");
                //Hide badge during (state chnage) animation
                document.querySelectorAll("#"+tgshId+"-tmw .avatar-badge").forEach(item => {
                    item.classList.add("displaynoneimportant");
                });
            }

            document.querySelectorAll("#"+animatedElementId+" .text-primary").forEach(item => {
                item.classList.add("animated-border-box");
            });

            let xhr = new XMLHttpRequest();
            xhrRequests.set(id,xhr);
            xhr.open('GET', '/action/'+id+"?otsseid="+ot_sse_id, true);
            xhr.onreadystatechange = function(e) {
                xhrRequests.delete(id);
                if (xhr.readyState == 4 && xhr.status == 200) {
                    if(xhr.responseText == "")
                        return;
                    let js = JSON.parse(xhr.responseText);
                    if(js["result"] == "ok") {
                        if(e.preventDefault)
                            e.preventDefault();
                        for(i = 0; i < js["cmds"].length ; i++)
                            runCommand(js["cmds"][i]);
                    }
                }
            };
            xhr.send();
        }, false);

        allJsActionButton[i].classList.add('jsact-processed');
    }
}

function sendRequestForPanelId(id,reqSuffix,param) {
    let xhr = new XMLHttpRequest();
    xhrRequests.set(id,xhr);
    let url_to_call = '/action/'+id+"-"+reqSuffix+"?otsseid="+ot_sse_id;
    if(param != null && param != "" && param.length > 0 )
        url_to_call += "&" + param;
    xhr.open('GET',url_to_call , true);
    xhr.onreadystatechange = function(e) {
        xhrRequests.delete(id);
        if (xhr.readyState == 4 && xhr.status == 200) {
            if(xhr.responseText == "")
                return;
            let js = JSON.parse(xhr.responseText);
            if(js["result"] == "ok") {
                for(i = 0; i < js["cmds"].length ; i++)
                    runCommand(js["cmds"][i]);
            }
        }
    };
    xhr.send();
}

function handleJustMovingPanels() {
    const allJustMoveButton = document.getElementsByClassName("justmove");
    for (let i = 0; i < allJustMoveButton.length; i++) {
        if(allJustMoveButton[i].classList.contains('justmove-processed'))
            continue;
        let grpId = allJustMoveButton[i].dataset.grpid;
        sendRequestForPanelId(grpId,"movupdate","");
        allJustMoveButton[i].classList.add('justmove-processed');
    }
}

function handleDeferredInfos() {
    const allNoInfoPanel = document.getElementsByClassName("panelnoinfo");
    for (let i = 0; i < allNoInfoPanel.length; i++) {
        if(allNoInfoPanel[i].classList.contains('panelnoinfo-processed'))
            continue;
        let refId = allNoInfoPanel[i].dataset.refid;
        sendRequestForPanelId(refId,"update","");
        allNoInfoPanel[i].classList.add('panelnoinfo-processed');
    }
}

function startSSE() {
    if(conf_use_sse === undefined || conf_use_sse != 1)
        return;

    let sseurl = window.location.protocol + "//" + window.location.hostname + ":" +
                 conf_sse_port + "/sse?id="+ot_sse_id+"&subscribe=thermostat-sensors-panelupd";
    const eventSource = new EventSource(sseurl);
    eventSource.onmessage = function(event) {
        if(event.data == "Hello")
            return;
        if(event.data == "SensorDataChanged") {
            const allSensorPanel = document.getElementsByClassName("sensorpanel");
            for (let i = 0; i < allSensorPanel.length; i++) {
                let refId = allSensorPanel[i].dataset.refid;
                sendRequestForPanelId(refId,"update","");
            }
        }
        if(event.data == "ThermostatDataChanged") {
            if(skip_therm_sse_requests)
                return;
            const allThermostatPanel = document.getElementsByClassName("thermostatpanel");
            for (let i = 0; i < allThermostatPanel.length; i++) {
                let refId = allThermostatPanel[i].dataset.refid;
                sendRequestForPanelId(refId,"update","");
            }
        }

        if(event.data.length > 11 && 
           event.data.substr(0,10) == "refreshId(" && 
           event.data.substr(event.data.length-1,1) == ")") {
            let ids = event.data.substr(10,event.data.length - 11)
            let idarray = ids.split(",");
            for(i = 0;i < idarray.length;i++) {
                let panel = document.getElementById("pc-" + idarray[i]);
                if(panel !== undefined && panel != null)
                    sendRequestForPanelId("b-" + idarray[i],"update","");
            }
        }
    };
}

function initUnravedGauges() {
    const allNoInfoPanel = document.getElementsByClassName("gauge-uncarved");
    for (let i = 0; i < allNoInfoPanel.length; i++) {
        let mainId = allNoInfoPanel[i].dataset.mid;
        let val = parseFloat(allNoInfoPanel[i].innerHTML);
        setTemperature(val,mainId);
    }
}

function setTemperature(temp,mainId) {
    const minTemp = 5;
    const maxTemp = 30;

    if (temp < minTemp) temp = minTemp;
    if (temp > maxTemp) temp = maxTemp;

    const percentage = (temp - minTemp) / (maxTemp - minTemp);
    const angle = 180 * percentage;

    const radius = 40;
    const x = 10 + radius + radius * Math.cos(Math.PI * (1 - percentage));
    const y = 50 - radius * Math.sin(Math.PI * (1 - percentage));

    const largeArcFlag = angle > 180 ? 1 : 0;
    const arcPath = `M 10 50 A 40 40 0 ${largeArcFlag} 1 ${x} ${y}`;
    document.getElementById('thermostatTempArc-'+mainId).setAttribute('d', arcPath);

    let color = 'white';
    for (let i = 0; i < thermostatTempColors.length - 1; i++) {
        if (temp >= thermostatTempColors[i].temp && temp < thermostatTempColors[i + 1].temp) {
            color = thermostatTempColors[i].color;
            break;
        }
    }

    document.getElementById('thermostatTempArc-'+mainId).setAttribute('stroke', color);
    let dispElement = document.getElementById('thermostatTemperatureDisplay-'+mainId);
    dispElement.textContent =
        (new Intl.NumberFormat('default', {
            minimumFractionDigits: 1,
            maximumFractionDigits: 1,
        })).format(temp).replace(",",".");
    dispElement.style.color = color;
    dispElement.dataset.fragval = Math.floor((temp - 5.0) * 2);
    dispElement.classList.remove('gauge-uncarved');
}

/* Clock picker stuff */
function clockbutton_spinner_handle_click(id,mainId) {
    let baseElement = document.getElementById(id);
    let targetId = baseElement.dataset.sid;
    let ival = parseInt(document.getElementById(targetId).innerHTML);
    let maxval = 0;
    if(baseElement.classList.contains("type-hour"))
        maxval = 23;
    if(baseElement.classList.contains("type-min"))
        maxval = 59;
    if(baseElement.classList.contains("dir-up")) {
        ival += 1;
        if(ival > maxval)
            ival = 0;
    }
    if(baseElement.classList.contains("dir-down")) {
        ival -= 1;
        if(ival < 0)
            ival = maxval;
    }

    document.querySelectorAll('#'+mainId + ' #'+targetId).forEach(item => {
        item.innerHTML = (ival < 10 ? "0" : "") + ival.toString();
    });
    if(baseElement.classList.contains("type-hour"))
        document.querySelectorAll('#'+mainId + ' #hidedhour').forEach(item => {
            item.value = ival.toString();
        });
    if(baseElement.classList.contains("type-min"))
        document.querySelectorAll('#'+mainId + ' #hidedmin').forEach(item => {
            item.value = ival.toString();
        });
}

function clockbutton_value_handle_click(overlay_container_id,mainId,targetId,hm) {
    let ophh = "<div>";
    ophh += "<table>";
    max = 0;
    if(hm == "hour")
        max = 23;
    if(hm == "min")
        max = 59;
    for(i = 0 ; i <= max ; i++) {
        ophh += "<tr><td><div class=\"overlay-select-button\" " +
                    "data-value=\""+i.toString()+"\" "+
                    "data-targetid=\""+targetId+"\">"+
                      (i < 10 ? "0" : "") + i.toString() +
                "</div></td></tr>";
    }
    ophh += "</div>";

    document.getElementById(overlay_container_id).innerHTML = ophh;
    document.getElementById(overlay_container_id).style.display = "block";

    const allOlButton = document.getElementsByClassName("overlay-select-button");
    for (let i = 0; i < allOlButton.length; i++) {
        let id = allOlButton[i].id;
        if(!allOlButton[i].classList.contains('overlay-select-button-processed')) {
            allOlButton[i].addEventListener("click", function(e) {
                let ival = parseInt(e.target.dataset.value);

                document.querySelectorAll('#'+mainId + ' #'+targetId).forEach(item => {
                    item.innerHTML = (ival < 10 ? "0" : "") + ival.toString();
                });
                if(hm == "hour")
                    document.querySelectorAll('#'+mainId + ' #hidedhour').forEach(item => {
                        item.value = ival.toString();
                    });
                if(hm == "min")
                    document.querySelectorAll('#'+mainId + ' #hidedmin').forEach(item => {
                        item.value = ival.toString();
                    });

                document.getElementById(overlay_container_id).innerHTML = "";
                document.getElementById(overlay_container_id).style.display = "none";
                e.preventDefault();
                clockselector_send_changed_if_required(mainId);
            });
            allOlButton[i].classList.add('overlay-select-button-processed');
        }
    }
}

function clockselector_send_changed_if_required(mainId) {
    let me = document.getElementById(mainId);
    if(me.classList.contains("jsfiredcs")) {
        let panelId = me.dataset.pnlid;
        let hourtxt = "";
        let mintxt = "";
        document.querySelectorAll('#'+mainId + ' #hidedhour').forEach(item => {
            hourtxt = item.value;
        });
        document.querySelectorAll('#'+mainId + ' #hidedmin').forEach(item => {
            mintxt = item.value;
        });
        sendRequestForPanelId("b-" + panelId,"updateclock","toclock=h" + hourtxt + "m" + mintxt);
    }
}

function init_clockselector(mainId) {
    document.querySelectorAll('#' + mainId +' .clock-spinner').forEach(item => {
        let id = item.id;
        if(!item.classList.contains('clock-spinner-processed')) {
            item.addEventListener("click", function(e) {
                clockbutton_spinner_handle_click(id,mainId);
                e.preventDefault();
            });
            item.classList.add('clock-spinner-processed');
        }
    });

    document.querySelectorAll('#' + mainId +' #t-hour').forEach(item => {
        item.addEventListener("click", function(e) {
            clockbutton_value_handle_click("overlay-pholder",mainId,"t-hour","hour");
            e.preventDefault();
        });
    });

    document.querySelectorAll('#' + mainId +' #t-min').forEach(item => {
        item.addEventListener("click", function(e) {
            clockbutton_value_handle_click("overlay-pholder",mainId,"t-min","min");
            e.preventDefault();
        });
    });
}

function initClockPickerBlocks() {
    const allClockPickerBlock = document.getElementsByClassName("clockpicker-controller-block");
    for (let i = 0; i < allClockPickerBlock.length; i++) {
        if(allClockPickerBlock[i].classList.contains('clockpicker-processed'))
            continue;
        let mainId = allClockPickerBlock[i].dataset.mainid;
        init_clockselector(mainId);
        allClockPickerBlock[i].classList.add('clockpicker-processed');
    }
}

function fillActionSubselect(main_select_value,subselect_id) {
    if(main_select_value.substr(0,7) == "switch:") {
        document.getElementById(subselect_id).innerHTML =
            '<option value="on">Switch On</option>'+
            '<option value="off">Switch Off</option>';
    }
    if(main_select_value.substr(0,7) == "action:") {
        document.getElementById(subselect_id).innerHTML =
            '<option value="run">Run</option>';
    }
    if(main_select_value.substr(0,8) == "shading:") {
        document.getElementById(subselect_id).innerHTML =
            '<option value="open">Open</option>'+
            '<option value="close">Close</option>';
    }
    if(main_select_value.substr(0,7) == "script:") {
        document.getElementById(subselect_id).innerHTML =
            '<option value="start">Start</option>'+
            '<option value="stop">Stop</option>';
    }
    if(main_select_value.substr(0,6) == "therm:") {
        let str;
        for(t=5.0;t<=30;t+=0.5) {
            str += "<option value=\"" + t.toString() + "\">" + t.toString() + "</option>";
        }
        document.getElementById(subselect_id).innerHTML = str;
    }
}

function initActionSubselector() {
    const allActionSelector = document.getElementsByClassName("schedule-action-selector");
    for (let i = 0; i < allActionSelector.length; i++) {
        if(allActionSelector[i].classList.contains('action-selector-processed'))
            continue;
        let actionSubId = allActionSelector[i].dataset.actionsubid;
        allActionSelector[i].addEventListener('change',function(e){
            fillActionSubselect(e.target.value,actionSubId);
        });
        allActionSelector[i].classList.add('action-selector-processed');
    }
}

function updateTime() {
    const today = new Date();
    let h = today.getHours();
    let m = today.getMinutes();
    let s = today.getSeconds();
    document.getElementById('tlclock').innerHTML =  (h < 10 ? "0" : "") + h + ":" + (m < 10 ? "0" : "") + m ;
    setTimeout(updateTime, 10000);
}

document.addEventListener("DOMContentLoaded", function(event) {
    xhrRequests.clear();
    initializeActions();
    handleJustMovingPanels();
    handleDeferredInfos();
    initUnravedGauges();
    initClockPickerBlocks();
    startSSE();
    updateTime();
});
