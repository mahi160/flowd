import { GitBranch } from "lucide-react";
import { Card } from "../components/card";
import { useData } from "../context";

export function RepoWidget() {
  const { data } = useData();

  return (
    <Card heading="Top repo" description="">
      <p className="font-display text-2xl font-light">{data?.top_repo}</p>
      <span className="summary-widget-badge">
        <GitBranch size={12} />
        {data?.top_branch}
      </span>
    </Card>
  );
}
