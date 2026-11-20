import { CanvasWorkspace } from "@/features/canvas/canvas-workspace";

export const instant = false;

export default async function ProjectPage({ params }: { params: Promise<{ projectId: string }> }) {
  const { projectId } = await params;
  return <CanvasWorkspace projectId={projectId} />;
}
