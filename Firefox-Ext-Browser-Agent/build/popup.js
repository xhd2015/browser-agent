// Firefox popup shell (P1). React popup integration can come later.
(function () {
  var hint = document.getElementById("hint");
  if (!hint) return;
  fetch("http://127.0.0.1:43761/v1/health")
    .then(function (r) {
      if (!r.ok) throw new Error("health " + r.status);
      return r.json();
    })
    .then(function () {
      hint.textContent =
        "Control server is reachable. Keep the temporary add-on loaded (reloads on Firefox restart).";
    })
    .catch(function () {
      hint.textContent =
        "Start browser-agent serve, then reload this temporary add-on if needed.";
    });
})();
