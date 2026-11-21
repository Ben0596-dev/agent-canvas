import { applyCanvasOperations } from "../canvas/reducer";
import type { CanvasSnapshot } from "../canvas/types";

import type { AgentProposal } from "./types";

export function previewProposal(snapshot: CanvasSnapshot, proposal: AgentProposal): CanvasSnapshot {
  return applyCanvasOperations(snapshot, proposal.operations);
}
