
var reqDataTimer = 3000
var nElementsAdded = 4
var downloadXSeconds = 7;
var KBYTE = 1024
var TICKINTERVAL = 1000
var XAXISRANGE = 4 * TICKINTERVAL
var speed = 1000 / TICKINTERVAL
var obj = { x: 0 }
var range = {
  min: -2,
  max: 2,
}



// Append new sample.
function append(series, buffer) {
  for (let j = 0; j < nElementsAdded; j++) {
    let now = Date.now();
    if(buffer.length == 0)
      return;
    let data = buffer.shift().split(",")
    //console.log(data)
    if (data.length < 4) {
      console.log("Error: " + data + buffer.length)
    }
    for (let i = 0; i < 4; i++) {
      //let value = Math.random() * (range.max - range.min) + range.min;
      // The append method takes a timestamp and a value.
      series[i].append(now, parseFloat(data[i]));
    }
  }

}

// Append new sample.
function zero(series) {
  let now = Date.now();
  for (let i = 0; i < 4; i++) {

    // The append method takes a timestamp and a value.
    series[i].append(now, 0);

  }

}

function downloadData(buffer, emotion, size, block, id, map) {
  data = `{"Id":"` + id + `", "Timestamp": "` + map.get("timestamp") + `","Emotion":"` + emotion + `","Size":` + size + `,"Offset":` + map.get("block") + `}`
  console.log(map.get("timestamp"))
  $.post("/getData", data, function (result) {
    obj = JSON.parse(result)
    console.log(obj)
    if (map.get("timestamp") == 0)
      map.set("timestamp", obj.Timestamp)
    buffer.push(...obj.Data)
    map.set("block",map.get("block")+1)

  });


}