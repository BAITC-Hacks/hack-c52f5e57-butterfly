// Reproducible API regression evaluation; synthetic fixtures do not measure AI accuracy.
import assert from 'node:assert/strict';
import { readFile, mkdir, writeFile } from 'node:fs/promises';
import { dirname, resolve } from 'node:path';
import { execFileSync } from 'node:child_process';

const args = process.argv.slice(2);
function option(name, fallback) { const i=args.indexOf(name); return i<0 ? fallback : args[i+1]; }
if (args.includes('--help')) {
  console.log('node scripts/eval.mjs [--base http://127.0.0.1:8000] [--out data/results/eval.json]\nRun only against a disposable test instance. Creates and deletes synthetic meetings. Never sends Telegram messages.');
  process.exit(0);
}
const base = option('--base', process.env.BUTTERFLY_TEST_URL || 'http://127.0.0.1:8000').replace(/\/$/,'');
const out = resolve(option('--out','data/results/eval.json'));
const fixture = JSON.parse(await readFile(new URL('../demo/eval/manual-unknown.json', import.meta.url)));
const checks=[], created=[];
let cookie='', stranger='', health={}, demo, manual;
const started = new Date().toISOString();
async function request(path, options={}, session=cookie) {
  return fetch(base+path,{...options,headers:{Cookie:session,...options.headers},signal:AbortSignal.timeout(25000)});
}
async function json(path,options={},session=cookie) {
  const response=await request(path,options,session);
  assert.ok(response.ok,`${path}: HTTP ${response.status} ${String(await response.clone().text()).slice(0,180)}`);
  return response.json();
}
function patch(value) { return {method:'PATCH',headers:{'Content-Type':'application/json'},body:JSON.stringify(value)}; }
async function check(name, run, condition=true) {
  if (!condition) { checks.push({name,status:'skipped',reason:'Required prerequisite or provider isolation unavailable'}); return; }
  const start=performance.now();
  try { await run(); checks.push({name,status:'passed',duration_ms:Math.round(performance.now()-start)}); }
  catch(error) { checks.push({name,status:'failed',duration_ms:Math.round(performance.now()-start),error:error.message}); }
}
await check('runtime and isolated sessions',async()=>{
  health=await json('/health'); assert.equal(health.runtime,'go');
  const one=await request('/api/session',{method:'POST'},''); assert.equal(one.status,200); cookie=one.headers.get('set-cookie').split(';')[0];
  const two=await request('/api/session',{method:'POST'},''); assert.equal(two.status,200); stranger=two.headers.get('set-cookie').split(';')[0]; assert.notEqual(cookie,stranger);
});
await check('synthetic demo is explicitly labelled',async()=>{
  demo=await json('/api/demo',{method:'POST'}); created.push(demo.id);
  assert.equal(demo.mode,'demo'); assert.equal(demo.status,'needs_review'); assert.equal(demo.actions.length,3);
  assert.ok(demo.events.some(e=>/Синтетический/.test(e.message)));
},!!cookie);
await check('source evidence is anchored to transcript',async()=>{
  for(const action of demo.actions) {
    const segment=demo.transcript.find(s=>s.id===action.segment_id); assert.ok(segment); assert.ok(segment.text.includes(action.evidence));
  }
  assert.ok(demo.actions.some(a=>a.owner===null&&a.deadline===null));
},!!demo);
await check('review gate blocks confirmation and PDF before review',async()=>{
  assert.equal((await request(`/api/meetings/${demo.id}/confirm`,{method:'POST'})).status,409);
  assert.equal((await request(`/api/meetings/${demo.id}/export?format=pdf`)).status,409);
},!!demo);
await check('cross-session isolation for read edit export delete',async()=>{
  const path=`/api/meetings/${demo.id}`;
  for(const [suffix,options] of [['',{}],['',patch({summary:'unauthorized'})],['/export?format=json',{}],['',{method:'DELETE'}]]) {
    assert.equal((await request(path+suffix,options,stranger)).status,404);
  }
  assert.ok(!(await json('/api/meetings',{},stranger)).meetings.some(m=>m.id===demo.id));
},!!demo&&!!stranger);
await check('review preserves unknown fields and evidence; confirm is idempotent',async()=>{
  const path=`/api/meetings/${demo.id}`;
  await json(path,patch({summary:demo.summary,actions:demo.actions.map(a=>({...a,status:'confirmed'}))}));
  const first=await json(path+'/confirm',{method:'POST'}); const second=await json(path+'/confirm',{method:'POST'});
  assert.equal(first.status,'completed'); assert.deepEqual(first.actions.map(a=>[a.owner,a.deadline,a.evidence]),demo.actions.map(a=>[a.owner,a.deadline,a.evidence]));
  assert.equal(first.events.length,second.events.length);
},!!demo);
for(const format of ['pdf','docx','json']) await check(`export ${format}`,async()=>{
  const response=await request(`/api/meetings/${demo.id}/export?format=${format}`); assert.equal(response.status,200);
  const bytes=Buffer.from(await response.arrayBuffer()); assert.ok(bytes.length>50);
  if(format==='pdf') assert.equal(bytes.subarray(0,4).toString(),'%PDF');
  if(format==='docx') assert.equal(bytes.subarray(0,2).toString(),'PK');
  if(format==='json') { const data=JSON.parse(bytes); assert.equal(data.status,'completed'); assert.ok(data.actions.some(a=>a.owner===null&&a.deadline===null)); }
},!!demo);
const isolated = !!cookie && health.provider?.llm_configured===false;
await check('missing provider retains manual transcript without invented actions',async()=>{
  const form=new FormData(); for(const key of ['title','language','date','transcript'])form.set(key,fixture[key]);
  manual=await json('/api/meetings',{method:'POST',body:form});created.push(manual.id);
  assert.equal(manual.mode,'manual'); assert.equal(manual.status,'needs_review'); assert.equal(manual.actions.length,0);
  assert.equal(manual.transcript.map(s=>s.text).join('\n'),fixture.transcript);
},isolated);
await check('manual unknown fields survive explicit human review and JSON export',async()=>{
  const path=`/api/meetings/${manual.id}`;
  await json(path,patch({summary:'Ручная проверка',actions:[{...fixture.action,segment_id:manual.transcript[0].id}]}));
  await json(path+'/confirm',{method:'POST'}); const exported=await json(path+'/export?format=json');
  assert.equal(exported.actions[0].owner,null);assert.equal(exported.actions[0].deadline,null);assert.equal(exported.actions[0].evidence,fixture.action.evidence);
},!!manual);
await check('missing ASR fails honestly without synthetic fallback',async()=>{
  const form=new FormData();form.set('title','Eval · unavailable ASR');form.set('file',new Blob([new Uint8Array([82,73,70,70,0,0,0,0,87,65,86,69])],{type:'audio/wav'}),'unavailable.wav');
  const meeting=await json('/api/meetings',{method:'POST',body:form});created.push(meeting.id);
  assert.equal(meeting.mode,'live');assert.equal(meeting.status,'awaiting_provider');assert.equal(meeting.actions.length,0);assert.equal(meeting.transcript.length,0);
},!!cookie&&health.provider?.asr_configured===false);
for(const id of created) await check(`cleanup ${id}`,async()=>assert.equal((await request(`/api/meetings/${id}`,{method:'DELETE'})).status,204));
let revision='unknown';try{revision=execFileSync('git',['rev-parse','HEAD'],{encoding:'utf8'}).trim();}catch{}
const counts=Object.fromEntries(['passed','failed','skipped'].map(status=>[status,checks.filter(c=>c.status===status).length]));
const report={schema_version:1,kind:'api_regression',started_at:started,finished_at:new Date().toISOString(),base_url:base,checkout_revision:revision,
  revision_note:'Client checkout revision; server revision is not attested by this script.',synthetic:true,counts,checks,
  model_quality:{status:'unverified',wer:null,action_precision:null,action_recall:null,reason:'No labelled real audio corpus or live ASR/LLM run in this regression suite.'},
  judge_round:{status:'not_run',valid_judges:0,reason:'Requires a frozen bundle and three isolated judges. No readiness or judging score is inferred.'}};
await mkdir(dirname(out),{recursive:true});await writeFile(out,JSON.stringify(report,null,2)+'\n');
console.log(JSON.stringify({report:out,...counts,model_quality:'unverified'},null,2));
process.exitCode=counts.failed?1:0;
