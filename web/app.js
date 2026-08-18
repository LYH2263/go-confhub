const $ = (id) => document.getElementById(id);
let currentNS = "";

async function api(path, opts) {
  const res = await fetch(path, {
    headers: { "Content-Type": "application/json", ...(opts && opts.headers) },
    ...opts,
  });
  const text = await res.text();
  let data = {};
  try { data = text ? JSON.parse(text) : {}; } catch (_) { data = { error: text }; }
  if (!res.ok) throw new Error(data.error || res.statusText);
  return data;
}

async function refreshNS() {
  const data = await api("/api/ns");
  const ul = $("ns-list");
  ul.innerHTML = "";
  (data.namespaces || []).forEach((n) => {
    const li = document.createElement("li");
    li.textContent = `${n.id}  (${n.owner})`;
    if (n.id === currentNS) li.classList.add("active");
    li.onclick = () => selectNS(n.id);
    ul.appendChild(li);
  });
}

async function selectNS(id) {
  currentNS = id;
  $("current-ns").textContent = "当前命名空间：" + id;
  await refreshNS();
  await refreshKeys();
  await refreshAudit();
}

async function refreshKeys() {
  const ul = $("key-list");
  ul.innerHTML = "";
  if (!currentNS) return;
  const data = await api(`/api/ns/${encodeURIComponent(currentNS)}/keys`);
  (data.keys || []).forEach((k) => {
    const li = document.createElement("li");
    li.innerHTML = `<strong>${k.key}</strong> <span class="muted">head=r${k.head} stable=r${k.stable}</span>`;
    li.onclick = async () => {
      const v = await api(`/api/ns/${encodeURIComponent(currentNS)}/keys/${encodeURIComponent(k.key)}`);
      $("key-id").value = k.key;
      $("payload").value = new TextDecoder().decode(
        Uint8Array.from(atob(typeof v.payload === "string" ? v.payload : ""), (c) => c.charCodeAt(0))
      ).replace(/\0/g, "") || "";
      try {
        $("payload").value = Array.isArray(v.payload)
          ? new TextDecoder().decode(new Uint8Array(v.payload))
          : (typeof v.payload === "string" ? v.payload : JSON.stringify(v.payload));
      } catch (_) {}
    };
    ul.appendChild(li);
  });
}

async function refreshAudit() {
  const data = await api("/api/audit?limit=20");
  const ul = $("audit-list");
  ul.innerHTML = "";
  (data.audit || []).forEach((a) => {
    const li = document.createElement("li");
    li.innerHTML = `<span class="muted">#${a.seq}</span> ${a.kind} ${a.ns}/${a.key || ""} r${a.rev || 0} ${a.ok ? "ok" : a.err}`;
    ul.appendChild(li);
  });
}

$("ns-form").addEventListener("submit", async (e) => {
  e.preventDefault();
  await api("/api/ns", {
    method: "POST",
    body: JSON.stringify({
      id: $("ns-id").value.trim(),
      owner: $("ns-owner").value.trim(),
      max_keys: Number($("ns-maxkeys").value || 0),
    }),
  });
  $("ns-id").value = "";
  await refreshNS();
});

$("put-form").addEventListener("submit", async (e) => {
  e.preventDefault();
  if (!currentNS) {
    alert("请先选择命名空间");
    return;
  }
  const pct = Number($("gray-pct").value || 0);
  const body = {
    payload: $("payload").value,
    author: $("author").value.trim(),
  };
  if (pct > 0) body.gray = { percent: pct };
  await api(`/api/ns/${encodeURIComponent(currentNS)}/keys/${encodeURIComponent($("key-id").value.trim())}`, {
    method: "PUT",
    body: JSON.stringify(body),
  });
  await refreshKeys();
  await refreshAudit();
});

refreshNS().then(refreshAudit).catch((err) => {
  $("audit-list").innerHTML = `<li>无法连接 API：${err.message}</li>`;
});
