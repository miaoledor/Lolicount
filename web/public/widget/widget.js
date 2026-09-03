/*!
 * Lolicount emote widget — renders an E-mote (PSB) character playing a
 * RANDOM motion on every page load, plus the visit count.
 *
 * Usage:
 *   <div data-lolicount="counter-name"
 *        data-model="azuki"
 *        data-text="you are the {n}-th visitor"
 *        data-width="320" data-height="480" data-fsize="18"></div>
 *   <script src="https://<host>/widget/widget.js" defer></script>
 *
 * Attributes:
 *   data-lolicount  counter name (required) — incremented on load
 *   data-model      emote model name (optional — first server model)
 *   data-text       count template, "{n}" is replaced by the count
 *   data-width      canvas CSS width in px (default 320)
 *   data-height     canvas CSS height in px (default 480)
 *   data-fsize      count text font size in px (default 18)
 *
 * Design doc: docs/emote-widget.md
 */
(function () {
  'use strict';

  // Resolve the server root from this script's own URL so the widget
  // works when embedded from any origin (.../widget/widget.js -> .../).
  var scriptEl = document.currentScript;
  var baseUrl = (scriptEl && scriptEl.src)
    ? new URL('../', scriptEl.src).href
    : '/';

  function loadScript(src) {
    return new Promise(function (resolve, reject) {
      var s = document.createElement('script');
      s.src = src;
      s.onload = function () { resolve(); };
      s.onerror = function () { reject(new Error('failed to load ' + src)); };
      document.head.appendChild(s);
    });
  }

  // The driver pair is loaded once per page and shared by all widget
  // instances. Order matters: emoteplayer.js declares the classes,
  // FreeMoteDriver.js provides the emscripten runtime they call into.
  var driverPromise = null;
  function ensureDriver() {
    if (!driverPromise) {
      driverPromise = loadScript(baseUrl + 'widget/emoteplayer.js')
        .then(function () { return loadScript(baseUrl + 'widget/FreeMoteDriver.js'); });
    }
    return driverPromise;
  }

  function webglAvailable() {
    try {
      var c = document.createElement('canvas');
      return !!(window.WebGLRenderingContext &&
        (c.getContext('webgl') || c.getContext('experimental-webgl')));
    } catch (e) {
      return false;
    }
  }

  function showError(el, msg) {
    var div = document.createElement('div');
    div.style.cssText = 'padding:8px 12px;font:13px/1.5 sans-serif;color:#b91c1c;' +
      'background:#fef2f2;border:1px solid #fecaca;border-radius:8px;';
    div.textContent = msg;
    el.appendChild(div);
  }

  function fetchCount(name) {
    return fetch(baseUrl + 'api/count/@' + encodeURIComponent(name), { credentials: 'omit' })
      .then(function (r) {
        if (!r.ok) throw new Error('count api error ' + r.status);
        return r.json();
      });
  }

  function renderCount(el, tpl, fsize, num) {
    var text = tpl ? tpl.split('{n}').join(String(num)) : String(num);
    var div = document.createElement('div');
    div.className = 'lolicount-emote-count';
    div.style.cssText = 'text-align:center;font-weight:600;margin-top:6px;';
    div.style.fontSize = fsize + 'px';
    div.textContent = text;
    el.appendChild(div);
  }

  // Scale and center the model inside the canvas (contain fit with 5%
  // padding), mirroring the reference autoCenterPlayer logic.
  function fitPlayer(player, canvas) {
    if (!player.isCharaProfileAvailable) return;
    var bounds = player.charaBounds;
    if (!bounds || bounds.right === bounds.left) return;
    var modelWidth = bounds.right - bounds.left;
    var modelHeight = bounds.bottom - bounds.top;
    if (modelWidth <= 0 || modelHeight <= 0) return;
    var scale = Math.min(canvas.width / modelWidth, canvas.height / modelHeight) * 0.95;
    var centerX = (bounds.left + bounds.right) / 2;
    var centerY = (bounds.top + bounds.bottom) / 2;
    player.setScale(scale, 0);
    player.setCoord(-centerX * scale, -centerY * scale, 0);
  }

// Labels that must never be picked at random: the "-----" separator rows
// some models embed in their label list, the initialization timeline, and
// gaze-follow (which needs pointer input to do anything visible).
var excludedMotion = function (label) {
  return label.indexOf('-') === 0 || label === '初期化' || label === '視線追従';
};

// Pick a random motion and play it. The engine's IsLoopTimeline query
// reports true for nearly every label right after a model load, so
// filtering by it is unreliable — instead startMotionLoop() keeps the
// character alive by re-picking whenever the canvas goes static.
function playRandomMotion(player, avoidLabel) {
  var labels = (player.mainTimelineLabels || []).filter(function (l) {
    return !excludedMotion(l) && l !== avoidLabel;
  });
  if (!labels.length) return null;
  var pick = labels[Math.floor(Math.random() * labels.length)];
  player.mainTimelineLabel = pick;
  return pick;
}

// Coarse strided sample of the visible canvas (luma+alpha byte per 64B),
// null when the canvas cannot be read.
function sampleCanvas(canvas) {
  try {
    var d = canvas.getContext('2d').getImageData(0, 0, canvas.width, canvas.height).data;
    var out = new Uint8Array(Math.ceil(d.length / 64));
    for (var i = 0, j = 0; i < d.length; i += 64, j++) {
      out[j] = (d[i] + d[i + 3]) & 0xff;
    }
    return out;
  } catch (e) {
    return null;
  }
}

// Keep the character animating: one-shot motions freeze on their last
// frame, so when the canvas stays (visibly) static for a few seconds,
// play another random motion. Sub-visual pixel noise is ignored via a
// per-sample threshold — a raw hash would never look "static".
function startMotionLoop(player, canvas) {
  var prev = null;
  var lastChange = Date.now();
  setInterval(function () {
    if (document.visibilityState !== 'visible') return;
    var cur = sampleCanvas(canvas);
    if (!cur) return;
    if (prev && cur.length === prev.length) {
      var changed = 0;
      for (var i = 0; i < cur.length; i++) {
        if (Math.abs(cur[i] - prev[i]) > 12) changed++;
      }
      // Real motion changes hundreds of samples; post-motion physics
      // settle and render noise stay below ~150. Anything above 50
      // counts as animating.
      if (changed > 50) {
        lastChange = Date.now();
      } else if (Date.now() - lastChange > 3000) {
        playRandomMotion(player, player._mainTimelineLabel);
        lastChange = Date.now();
      }
    }
    prev = cur;
  }, 900);
}

  function resolveModel(requested) {
    if (requested) return Promise.resolve(requested);
    return fetch(baseUrl + 'api/psb/models', { credentials: 'omit' })
      .then(function (r) { return r.json(); })
      .then(function (d) {
        var models = (d && d.models) || [];
        if (!models.length) throw new Error('no emote models available on this server');
        return models[0];
      });
  }

  function initWidget(el) {
    var name = el.getAttribute('data-lolicount');
    if (!name) return;

    var tpl = el.getAttribute('data-text') || '';
    var fsize = parseInt(el.getAttribute('data-fsize') || '18', 10) || 18;
    var cssW = parseInt(el.getAttribute('data-width') || '320', 10) || 320;
    var cssH = parseInt(el.getAttribute('data-height') || '480', 10) || 480;

    if (!webglAvailable()) {
      showError(el, 'lolicount emote widget: WebGL is not available in this browser');
      return;
    }

    ensureDriver()
      .then(function () {
        if (typeof EmotePlayer === 'undefined') {
          throw new Error('emote driver failed to initialize');
        }
        return resolveModel(el.getAttribute('data-model'));
      })
      .then(function (model) {
        var dpr = Math.min(window.devicePixelRatio || 1, 2);
        var canvas = document.createElement('canvas');
        canvas.width = Math.round(cssW * dpr);
        canvas.height = Math.round(cssH * dpr);
        canvas.style.width = cssW + 'px';
        canvas.style.height = cssH + 'px';
        canvas.style.display = 'block';
        canvas.style.margin = '0 auto';
        el.appendChild(canvas);

        EmotePlayer.createRenderCanvas(canvas.width, canvas.height);
        var player = new EmotePlayer(canvas);
        return player.promiseLoadDataFromURL(baseUrl + 'psb/' + encodeURIComponent(model))
          .then(function () {
            fitPlayer(player, canvas);
            playRandomMotion(player);
            startMotionLoop(player, canvas);
          });
      })
      .then(function () {
        return fetchCount(name);
      })
      .then(function (d) {
        renderCount(el, tpl, fsize, d.num);
      })
      .catch(function (err) {
        console.error('[lolicount-emote]', err);
        showError(el, 'lolicount emote widget: ' + (err && err.message ? err.message : 'load failed'));
      });
  }

  function boot() {
    var nodes = document.querySelectorAll('[data-lolicount]');
    Array.prototype.forEach.call(nodes, initWidget);
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', boot);
  } else {
    boot();
  }
})();
