import { FC, useEffect, useRef } from 'react';
import prism from 'prismjs';

// Add river definitions to Prism. We only define elements we expliclty have
// syntax highlighting rules for.
prism.languages.river = {
  blockHeader: {
    pattern: /^\s*[^=]+{/m,
    inside: {
      selector: {
        pattern: /([A-Za-z_][A-Za-z0-9_]*)(.([A-Za-z_][A-Za-z0-9_]*))*/,
      },
      comment: {
        pattern: /\/\/.*|\/\*[\s\S]*?(?:\*\/|$)/,
        greedy: true,
      },
      string: {
        pattern: /(^|[^\\])"(?:\\.|[^\\"\r\n])*"(?!\s*:)/,
        lookbehind: true,
        greedy: true,
      },
    },
  },
  comment: {
    pattern: /\/\/.*|\/\*[\s\S]*?(?:\*\/|$)/,
    greedy: true,
  },
  number: /-?\b\d+(?:\.\d+)?(?:e[+-]?\d+)?\b/i,
  string: {
    pattern: /(^|[^\\])"(?:\\.|[^\\"\r\n])*"(?!\s*:)/,
    lookbehind: true,
    greedy: true,
  },
  boolean: /\b(?:false|true)\b/,
  null: {
    pattern: /\bnull\b/,
    alias: 'keyword',
  },
};

/**
 * RiverBlob renders text as syntax highlighted River code.
 */
export const RiverBlob: FC<{ children: string }> = ({ children }) => {
  const codeRef = useRef<HTMLPreElement>(null);

  useEffect(() => {
    if (codeRef.current == null) {
      return;
    }

    prism.highlightAllUnder(codeRef.current);
  }, []);

  return (
    <pre ref={codeRef} style={{ margin: '0px', fontSize: '14px' }}>
      <code className="language-river">{children}</code>
    </pre>
  );
};
