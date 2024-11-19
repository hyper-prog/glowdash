/*
 Notify glowdash when input changed (wall switch action)
*/

let glowdashaddress = "192.168.1.100"
let userdata = null;

function notifyGlowdash() {
  Shelly.call(
    "HTTP.GET", {
          url: "http://" + glowdashaddress + "/hit",
          timeout: 1
    },
    function(res, error_code, error_msg) {}
  );
}

function btncallback(userdata) {
  if(userdata.info.component.substr(0,6) == "input:" && userdata.info.event == "toggle") {
    notifyGlowdash();
  }
}

Shelly.addEventHandler(btncallback,userdata);
