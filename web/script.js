let canvasElements = [];

function fetchCanvasLines() {
  fetch("/canvas/lines")
    .then((response) => {
      if (!response.ok) {
        throw new Error("Network response was not ok " + response.statusText);
      }
      return response.json();
    })
    .then((data) => {
      if (data.lines != null) {
        canvasElements = data.lines;
      }
      draw();
    })
    .catch((error) => {
      console.error("Could not fetch canvas lines:", error);
    });
}

fetchCanvasLines();

// prepare canvas and make it cover entire screen
const canvas = document.getElementById("canvas");
const ctx = canvas.getContext("2d");
canvas.width = window.innerWidth;
canvas.height = window.innerHeight;

addEventListener("resize", (event) => {
  // resize
  canvas.width = window.innerWidth;
  canvas.height = window.innerHeight;

  // resizing clears screen and so we need to redraw
  draw();
});

// clear screen and update screen according to stored canvas elements
function draw() {
  if (canvasElements == null) {
    return;
  }

  // clear screen
  ctx.clearRect(0, 0, canvas.width, canvas.height);

  // apply the accumulated translation
  ctx.save();
  ctx.translate(xTranslate, yTranslate);

  // draw canvas elements
  canvasElements.forEach((line) => {
    let pointPrev = line.points[0];
    ctx.beginPath();
    ctx.strokeStyle = line.color;
    line.points.forEach((point) => {
      ctx.moveTo(pointPrev.x, pointPrev.y);
      ctx.lineTo(point.x, point.y);
      pointPrev = point;
      ctx.stroke();
    });
  });

  ctx.restore();
}

// websocket handling
const wsUrl =
  (window.location.protocol === "https:" ? "wss:" : "ws:") +
  "//" +
  window.location.host +
  "/ws";
let ws = new WebSocket(wsUrl);
ws.onmessage = (event) => {
  canvasElements.push(JSON.parse(event.data).data);
  draw();
};

// handling of pen vs hand mode
let mode;

function handMode() {
  document.getElementById("penNav").classList.remove("active");
  document.getElementById("handNav").classList.add("active");
  mode = "hand";
  canvas.style.cursor = "grab";
}

function penMode() {
  document.getElementById("penNav").classList.add("active");
  document.getElementById("handNav").classList.remove("active");
  mode = "pen";
  canvas.style.cursor = "default";
}

penMode(); // start in pen mode

// color picker logic
const colorPicker = document.getElementById("color-picker");
let currentColor = colorPicker.value;
colorPicker.addEventListener("change", (event) => {
  currentColor = event.target.value;
});

// draw and drag logic
let isMouseDown = false;
let xPrev = 0;
let yPrev = 0;
let xTranslate = 0;
let yTranslate = 0;

canvas.addEventListener("mousedown", (event) => {
  if (event.button != 0) {
    return; // only left mouse button
  }
  isMouseDown = true;
  xPrev = event.offsetX;
  yPrev = event.offsetY;

  // cursor style
  if (mode == "hand") {
    canvas.style.cursor = "grabbing";
  }
});

canvas.addEventListener("mouseleave", (event) => {
  isMouseDown = false;

  // cursor style
  if (mode == "hand") {
    canvas.style.cursor = "grab";
  }
});

canvas.addEventListener("mouseup", (event) => {
  if (event.button != 0) {
    return; // only left mouse button
  }
  isMouseDown = false;

  // cursor style
  if (mode == "hand") {
    canvas.style.cursor = "grab";
  }
});

canvas.addEventListener("mousemove", (event) => {
  const xNew = event.offsetX;
  const yNew = event.offsetY;

  if (!isMouseDown) {
    return;
  }
  if (mode == "pen") {
    ws.send(
      JSON.stringify({
        type: "line",
        data: {
          points: [
            {
              x: xPrev - xTranslate,
              y: yPrev - yTranslate,
            },
            {
              x: xNew - xTranslate,
              y: yNew - yTranslate,
            },
          ],
          color: currentColor,
          lineWidth: 1,
        },
      })
    );
    xPrev = xNew;
    yPrev = yNew;
  }
  if (mode == "hand") {
    xTranslate += xNew - xPrev;
    yTranslate += yNew - yPrev;
    draw();
    xPrev = xNew;
    yPrev = yNew;
  }
});

function getPointerPosition(event) {
  if (event.touches && event.touches.length > 0) {
    return {
      x: event.touches[0].clientX - canvas.getBoundingClientRect().left,
      y: event.touches[0].clientY - canvas.getBoundingClientRect().top,
    };
  } else {
    return {
      x: event.offsetX,
      y: event.offsetY,
    };
  }
}

canvas.addEventListener("touchstart", (event) => {
  event.preventDefault();
  if (
    (event.button != undefined && event.button != 0) ||
    (event.touches && event.touches.length != 1)
  ) {
    return; // only single touch
  }
  isMouseDown = true;
  const { x, y } = getPointerPosition(event);
  xPrev = x;
  yPrev = y;
});

canvas.addEventListener("touchend", (event) => {
  event.preventDefault();
  if (
    (event.button != undefined && event.button != 0) ||
    (event.touches && event.touches.length != 1)
  ) {
    return; // only single touch
  }
  isMouseDown = false;
});

canvas.addEventListener("touchmove", (event) => {
  event.preventDefault();
  if (!isMouseDown) {
    return;
  }
  const { x: xNew, y: yNew } = getPointerPosition(event);

  if (mode == "pen") {
    ws.send(
      JSON.stringify({
        type: "line",
        data: {
          points: [
            {
              x: xPrev - xTranslate,
              y: yPrev - yTranslate,
            },
            {
              x: xNew - xTranslate,
              y: yNew - yTranslate,
            },
          ],
          color: currentColor,
          lineWidth: 1,
        },
      })
    );
    xPrev = xNew;
    yPrev = yNew;
  }
  if (mode == "hand") {
    xTranslate += xNew - xPrev;
    yTranslate += yNew - yPrev;
    draw();
    xPrev = xNew;
    yPrev = yNew;
  }
});

canvas.addEventListener("touchcancel", (event) => {
  event.preventDefault();
  isMouseDown = false;
});
