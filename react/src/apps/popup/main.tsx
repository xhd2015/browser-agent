import { createRoot } from "react-dom/client";
import { PopupApp } from "../../ui/PopupApp";
import { browserAgentProduct } from "../../products/browser-agent";

const el = document.getElementById("root") ?? document.body;
createRoot(el).render(<PopupApp product={browserAgentProduct} />);
