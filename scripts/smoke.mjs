// Run against a local running server. Uses synthetic data and sends no Telegram messages.
import assert from 'node:assert/strict';
const base = process.env.BUTTERFLY_TEST_URL || 'http://127.0.0.1:8000';
let cookie = '';
async function call(path, options={}) {
  const headers = {Cookie: cookie, ...options.headers};
  const response = await fetch(base+path, {...options,headers});
  return response;
}
const session = await call('/api/session',{method:'POST'});
assert.equal(session.status,200);
cookie = session.headers.get('set-cookie').split(';')[0];
const health = await (await call('/health')).json();
assert.equal(health.runtime,'go');
const demoResponse = await call('/api/demo',{method:'POST'});
assert.ok(demoResponse.ok,await demoResponse.clone().text());
let meeting = await demoResponse.json();
assert.equal(meeting.mode,'demo');
assert.ok(meeting.actions.length);
const url = '/api/meetings/'+meeting.id;
assert.equal((await call(url+'/export?format=pdf')).status,409);
const updated = await call(url,{method:'PATCH',headers:{'Content-Type':'application/json'},body:JSON.stringify({summary:meeting.summary,actions:meeting.actions.map(a=>({...a,status:'confirmed'}))})});
assert.ok(updated.ok,await updated.clone().text());
meeting=await (await call(url+'/confirm',{method:'POST'})).json();
assert.equal(meeting.status,'completed');
for (const format of ['pdf','docx','json']) {
  const response=await call(url+'/export?format='+format);
  assert.equal(response.status,200,await response.clone().text());
  const bytes=Buffer.from(await response.arrayBuffer());
  assert.ok(bytes.length>50);
  if(format==='pdf')assert.equal(bytes.subarray(0,4).toString(),'%PDF');
  if(format==='docx')assert.equal(bytes.subarray(0,2).toString(),'PK');
}
const stranger=await fetch(base+'/api/session',{method:'POST'});
const otherCookie=stranger.headers.get('set-cookie').split(';')[0];
assert.equal((await call(url,{headers:{Cookie:otherCookie}})).status,404);
if (health.mcp.fly_available) {
  const fly=await call('/fly-demo');assert.equal(fly.status,200);
  assert.match(await fly.text(),/MB 3D/);
  assert.equal((await fetch(base+'/fly-demo')).status,401);
}
assert.equal((await call('/api/meetings/'+meeting.id,{method:'DELETE'})).status,204);
console.log(JSON.stringify({status:'passed',scenario:'synthetic review -> PDF/DOCX/JSON + ownership + Fly',runtime:health.runtime,mcp:health.mcp},null,2));
