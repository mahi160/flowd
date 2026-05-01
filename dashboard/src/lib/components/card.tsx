import type { HTMLAttributes } from "preact";

export interface CardProps extends HTMLAttributes<HTMLDivElement> {
  header: string;
  description: string;
}
export function Card({ children, header, description, ...props }: CardProps) {
  return (
    <div
      className="rounded-xl shadow bg-surface p-4 flex flex-col gap-3"
      {...props}
    >
      <div className="flex gap-2 items-baseline justify-start">
        <h2 className="font-display text-xs tracking-wider font-light leading-none text-foreground/60 uppercase">
          {header}
        </h2>
        <h4 className="font-mono text-xs text-foreground/40">{description}</h4>
      </div>
      <div>{children}</div>
    </div>
  );
}
