import type { model } from "@gcsim/types";
import { Portraits } from "@gcsim/components/src/Cards/SatoriPreviewCard/Portraits";
import { TimelineChart } from "@gcsim/components/src/Cards/SatoriPreviewCard/charts/TimelineChart";
import { HistogramChart } from "@gcsim/components/src/Cards/SatoriPreviewCard/charts/HistogramChart";
import { CharacterPie, ElementPie } from "@gcsim/components/src/Cards/SatoriPreviewCard/charts/PieChart";
import { PORTRAIT_H, PORTRAIT_W } from "@gcsim/components/src/Cards/SatoriPreviewCard/portraitGeometry";

const number = new Intl.NumberFormat("en", { maximumFractionDigits: 0 });

// KQM presentation stays outside the upstream preview component.
export function Preview({ data }: { data: model.SimulationResult }) {
  const stats = data.statistics;
  const targets = data.target_details?.length ?? 1;
  const dps = (stats?.dps?.mean ?? 0) / Math.max(targets, 1);
  return <div className="preview-card">
    <Portraits data={data} width={PORTRAIT_W} height={PORTRAIT_H} margin={4} resolveAsset={path => `/api/assets/${path}`} />
    <div className="preview-summary">
      <strong className="preview-brand">KQM SIM</strong>
      <span><strong>{number.format(dps)}</strong> DPS / target</span>
      <span>{number.format(stats?.iterations ?? 0)} runs</span>
      <span>{number.format(stats?.duration?.mean ?? 0)}s</span>
    </div>
    <div className="preview-charts">
      <div><TimelineChart data={stats?.damage_buckets} width={154} height={84} /></div>
      <div><CharacterPie dps={stats?.character_dps} width={106} height={84} /></div>
      <div><ElementPie dps={stats?.element_dps} width={106} height={84} /></div>
      <div><HistogramChart data={stats?.dps} width={154} height={84} /></div>
    </div>
  </div>;
}
