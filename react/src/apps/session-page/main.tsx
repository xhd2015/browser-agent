import { createRoot } from "react-dom/client";
import { SessionPageApp } from "../../ui/SessionPageApp";
import { browserAgentProduct } from "../../products/browser-agent";

const el = document.getElementById("root") ?? document.body;
createRoot(el).render(
  <SessionPageApp product={browserAgentProduct} />
);
