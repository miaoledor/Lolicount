// Render at build time only. A matching Spine 3.6 runtime is supplied by the caller.
window.renderModel = async (model, maxEdge = 800) => {
  const canvas = document.querySelector('canvas');
  // Atlas alpha already contains antialiasing. MSAA would resolve overlapping
  // quads differently when exporting parts independently.
  const gl = canvas.getContext('webgl', {preserveDrawingBuffer: true, alpha: true, antialias: false});
  if (!gl) throw new Error('WebGL unavailable');
  // Spine 3.6 uses blendFunc for both channels, squaring straight alpha.
  // Separate alpha factors are required for portable transparent layers.
  gl.blendFunc = (source, destination) => gl.blendFuncSeparate(source, destination, gl.ONE, destination);
  const dir = `/models/${model}/`;
  const manager = new spine.webgl.AssetManager(gl);
  const data = await (await fetch(dir + 'dyn.json')).json();
  await new Promise((resolve, reject) => manager.loadTextureAtlas(dir + 'dyn.atlas', resolve, reject));
  const reader = new spine.SkeletonJson(new spine.AtlasAttachmentLoader(manager.get(dir + 'dyn.atlas')));
  const skeleton = new spine.Skeleton(reader.readSkeletonData(data));
  skeleton.setToSetupPose();
  skeleton.updateWorldTransform();
  const items = [];
  for (const skin of skeleton.data.skins) {
    for (let slot = 0; slot < skin.attachments.length; slot++) {
      for (const name of Object.keys(skin.attachments[slot] || {})) {
        const category = /^eye/.test(name) ? 'eye' : /^mouth/.test(name) ? 'mouth' : 'body';
        items.push({skin: skin.name, slot, name, category});
      }
    }
  }
  const show = (selected) => {
    for (const slot of skeleton.slots) slot.setAttachment(null);
    for (const item of selected) {
      skeleton.slots[item.slot].setAttachment(skeleton.data.findSkin(item.skin).getAttachment(item.slot, item.name));
    }
    skeleton.updateWorldTransform();
  };
  const bounds = [];
  for (const item of items) {
    show([item]);
    const offset = new spine.Vector2(), size = new spine.Vector2();
    skeleton.getBounds(offset, size, []);
    if (![offset.x, offset.y, size.x, size.y].every(Number.isFinite)) throw new Error('Invalid bounds');
    bounds.push([offset.x, offset.y, offset.x + size.x, offset.y + size.y]);
  }
  const x = Math.min(...bounds.map(b => b[0])), y = Math.min(...bounds.map(b => b[1]));
  const w = Math.max(...bounds.map(b => b[2])) - x, h = Math.max(...bounds.map(b => b[3])) - y;
  const scale = Math.min(1, maxEdge / Math.max(w, h));
  canvas.width = Math.ceil(w * scale) + 4;
  canvas.height = Math.ceil(h * scale) + 4;
  const renderer = new spine.webgl.SceneRenderer(canvas, gl, false);
  renderer.camera.zoom = 1 / scale;
  renderer.camera.position.x = x + w / 2;
  renderer.camera.position.y = y + h / 2;
  const render = (selected, crop = false) => {
    show(selected);
    gl.viewport(0, 0, canvas.width, canvas.height);
    gl.clearColor(0, 0, 0, 0);
    gl.clear(gl.COLOR_BUFFER_BIT);
    renderer.begin();
    renderer.drawSkeleton(skeleton, false);
    renderer.end();
    let output = canvas, left = 0, top = 0;
    if (crop) {
      const offset = new spine.Vector2(), size = new spine.Vector2();
      skeleton.getBounds(offset, size, []);
      left = Math.max(0, Math.floor((offset.x - renderer.camera.position.x) * scale + canvas.width / 2) - 2);
      top = Math.max(0, Math.floor((renderer.camera.position.y - offset.y - size.y) * scale + canvas.height / 2) - 2);
      const right = Math.min(canvas.width, Math.ceil((offset.x + size.x - renderer.camera.position.x) * scale + canvas.width / 2) + 2);
      const bottom = Math.min(canvas.height, Math.ceil((renderer.camera.position.y - offset.y) * scale + canvas.height / 2) + 2);
      output = document.createElement('canvas');
      output.width = Math.max(1, right - left);
      output.height = Math.max(1, bottom - top);
      output.getContext('2d').drawImage(canvas, left, top, output.width, output.height, 0, 0, output.width, output.height);
    }
    return {png: output.toDataURL('image/png').split(',')[1], left, top};
  };
  const bodies = new Map();
  for (const item of items.filter(i => i.category === 'body')) {
    if (!bodies.has(item.skin)) bodies.set(item.skin, []);
    bodies.get(item.skin).push(item);
  }
  const parts = [...bodies].map(([skin, body]) => ({category: 'body', name: `${skin}_body`, ...render(body, true)}));
  const categories = [...new Set(items.map(i => i.category))].sort((a, b) =>
    Math.min(...items.filter(i => i.category === a).map(i => i.slot)) -
    Math.min(...items.filter(i => i.category === b).map(i => i.slot)));
  for (const category of categories.filter(c => c !== 'body')) {
    for (const item of items.filter(i => i.category === category)) {
      parts.push({category: item.category, name: `${item.skin}_${item.name}`, ...render([item], true)});
    }
  }
  const references = [];
  for (const skin of skeleton.data.skins.filter(s => s.name !== 'default')) {
    const expression = items.filter(i => i.skin === skin.name);
    const eye = expression.find(i => i.name === 'eyeOpen');
    const mouth = expression.find(i => i.name === 'mouthClose');
    const selected = [eye, mouth].filter(Boolean);
    const bodySkin = bodies.has(skin.name) ? skin.name : 'default';
    references.push({skin: skin.name, body: `${bodySkin}_body`, names: selected.map(i => `${i.skin}_${i.name}`), png: render([...bodies.get(bodySkin), ...selected]).png});
  }
  manager.dispose();
  return {parts, references, width: canvas.width, height: canvas.height};
};
