"use client";

import { ChangeEvent, useCallback, useEffect, useMemo, useRef, useState } from "react";

type Selection = "automatic" | "included" | "excluded";
type Allocation = "split_evenly" | "participant_one" | "participant_two";

type Transaction = {
  id: string;
  date: string;
  description: string;
  amount_cents: number;
  source_file?: string;
  source_line?: number;
  selection: Selection;
  split_amount_cents?: number;
  allocation: Allocation;
};

type CalculationRow = Omit<Transaction, "selection"> & { split_amount_cents: number };

type Calculation = {
  included: CalculationRow[];
  unmatched: CalculationRow[];
  totals: {
    total_cents: number;
    participant_one_half_cents: number;
    participant_two_half_cents: number;
  };
};

type APIEnvelope<T> = {
  data?: T;
  error?: { code: string; message: string };
};

const fallbackPrefixes = ["HEMKOP", "ICA", "MAXI ICA", "WILLYS", "COOP", "PRESSBYRÅN"];

export function CostSplitter() {
  const [currency, setCurrency] = useState("SEK");
  const [prefixesText, setPrefixesText] = useState(fallbackPrefixes.join("\n"));
  const [files, setFiles] = useState<File[]>([]);
  const [fileCount, setFileCount] = useState(0);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const transactionsRef = useRef<Transaction[]>([]);
  const [calculation, setCalculation] = useState<Calculation | null>(null);
  const [amountDrafts, setAmountDrafts] = useState<Record<string, string>>({});
  const [showUnmatched, setShowUnmatched] = useState(false);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    void requestAPI<{ currency: string; prefixes: string[] }>("/backend/api/v1/defaults")
      .then((defaults) => {
        setCurrency(defaults.currency);
        setPrefixesText(defaults.prefixes.join("\n"));
      })
      .catch((reason: unknown) => setError(messageFrom(reason)));
  }, []);

  const prefixes = useMemo(
    () => prefixesText.split(/\r?\n/).map((value) => value.trim()).filter(Boolean),
    [prefixesText],
  );

  const calculate = useCallback(async (nextTransactions: Transaction[]) => {
    const result = await requestAPI<Calculation>("/backend/api/v1/split-calculations", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ prefixes, transactions: nextTransactions }),
    });
    setCalculation(result);
  }, [prefixes]);

  async function analyzeFiles() {
    if (files.length === 0) {
      setError("Choose at least one American Express CSV file.");
      return;
    }
    setBusy(true);
    setError("");
    try {
      const form = new FormData();
      files.forEach((file) => form.append("files", file));
      const imported = await requestAPI<{
        files_count: number;
        rows_count: number;
        transactions: Omit<Transaction, "selection" | "allocation">[];
      }>("/backend/api/v1/imports/amex", { method: "POST", body: form });
      const nextTransactions = imported.transactions.map<Transaction>((item) => ({
        ...item,
        selection: "automatic",
        allocation: "split_evenly",
      }));
      setTransactions(nextTransactions);
      transactionsRef.current = nextTransactions;
      setFileCount(imported.files_count);
      setAmountDrafts(Object.fromEntries(nextTransactions.map((item) => [item.id, amountInput(item.amount_cents)])));
      await calculate(nextTransactions);
    } catch (reason) {
      setError(messageFrom(reason));
    } finally {
      setBusy(false);
    }
  }

  async function recalculate() {
    setBusy(true);
    setError("");
    try {
      await calculate(transactionsRef.current);
    } catch (reason) {
      setError(messageFrom(reason));
    } finally {
      setBusy(false);
    }
  }

  async function updateTransaction(id: string, patch: Partial<Transaction>) {
    const nextTransactions = transactionsRef.current.map((item) => item.id === id ? { ...item, ...patch } : item);
    transactionsRef.current = nextTransactions;
    setTransactions(nextTransactions);
    setError("");
    try {
      await calculate(nextTransactions);
    } catch (reason) {
      setError(messageFrom(reason));
    }
  }

  async function setIncluded(id: string, included: boolean) {
    const transaction = transactionsRef.current.find((item) => item.id === id);
    if (!transaction) return;
    setAmountDrafts((drafts) => ({ ...drafts, [id]: amountInput(transaction.amount_cents) }));
    await updateTransaction(id, {
      selection: included ? "included" : "excluded",
      allocation: "split_evenly",
      split_amount_cents: undefined,
    });
  }

  function changeFiles(event: ChangeEvent<HTMLInputElement>) {
    setFiles(Array.from(event.target.files ?? []));
  }

  async function commitAmount(row: CalculationRow) {
    const value = amountDrafts[row.id] ?? amountInput(row.split_amount_cents);
    const parsed = parseAmountCents(value);
    if (parsed === null) {
      setError(`Amount to split for ${row.description} is not a valid monetary amount.`);
      return;
    }
    setAmountDrafts((drafts) => ({ ...drafts, [row.id]: amountInput(parsed) }));
    await updateTransaction(row.id, { split_amount_cents: parsed });
  }

  return (
    <>
      <header>
        <div className="wrap topbar">
          <div>
            <h1>Cost Splitter</h1>
            <p className="product-context">Shared expense workspace</p>
          </div>
          <span className="api-status">Next.js · Go API v1</span>
        </div>
      </header>

      <main className="wrap">
        {error && <div className="error" role="alert">{error}</div>}
        <div className="layout">
          <aside className="panel controls">
            <div className="panel-heading">
              <div>
                <h2>Import activity</h2>
                <p>Upload AmEx CSV exports and tune the matching rules.</p>
              </div>
            </div>
            <div className="form-grid">
              <label>
                CSV files
                <input type="file" accept=".csv,text/csv" multiple onChange={changeFiles} />
                <span className="hint">Files stay in memory and are not stored.</span>
              </label>
              <label>
                Currency
                <input value={currency} onChange={(event) => setCurrency(event.target.value.toUpperCase())} maxLength={8} />
              </label>
              <label>
                Merchant prefixes
                <textarea value={prefixesText} onChange={(event) => setPrefixesText(event.target.value)} />
              </label>
              <label className="check">
                <input type="checkbox" checked={showUnmatched} onChange={(event) => setShowUnmatched(event.target.checked)} />
                Show unmatched transactions
              </label>
              <button onClick={transactions.length ? recalculate : analyzeFiles} disabled={busy}>
                {busy ? "Working…" : transactions.length ? "Recalculate" : "Analyze CSV Files"}
              </button>
              {transactions.length > 0 && (
                <button className="secondary" onClick={analyzeFiles} disabled={busy || files.length === 0}>
                  Import selected files again
                </button>
              )}
            </div>
          </aside>

          <section className="panel results" aria-live="polite" aria-busy={busy}>
            {!calculation ? (
              <div className="empty-state">
                <h2>Review the split</h2>
                <p>Your included transactions, allocation controls, and participant totals will appear here.</p>
              </div>
            ) : (
              <>
                <div className="summary">
                  <Metric label="Files" value={String(fileCount)} />
                  <Metric label="Rows" value={String(transactions.length)} />
                  <Metric label="Included" value={String(calculation.included.length)} />
                  <Metric label="Total" value={formatCents(currency, calculation.totals.total_cents)} />
                  <Metric label="Participant one" value={formatHalfCents(currency, calculation.totals.participant_one_half_cents)} />
                  <Metric label="Participant two" value={formatHalfCents(currency, calculation.totals.participant_two_half_cents)} />
                </div>

                <TransactionTable
                  title="Included transactions"
                  rows={calculation.included}
                  currency={currency}
                  amountDrafts={amountDrafts}
                  setAmountDrafts={setAmountDrafts}
                  onCommitAmount={commitAmount}
                  onAllocation={(id, allocation) => updateTransaction(id, { allocation })}
                  onToggle={(id) => setIncluded(id, false)}
                  toggleLabel="Remove"
                  editable
                />

                {showUnmatched && (
                  <TransactionTable
                    title="Unmatched transactions"
                    rows={calculation.unmatched}
                    currency={currency}
                    amountDrafts={amountDrafts}
                    setAmountDrafts={setAmountDrafts}
                    onCommitAmount={commitAmount}
                    onAllocation={(id, allocation) => updateTransaction(id, { allocation })}
                    onToggle={(id) => setIncluded(id, true)}
                    toggleLabel="Include"
                    editable={false}
                  />
                )}
              </>
            )}
          </section>
        </div>
      </main>
    </>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return <div className="metric"><span>{label}</span><strong>{value}</strong></div>;
}

type TableProps = {
  title: string;
  rows: CalculationRow[];
  currency: string;
  amountDrafts: Record<string, string>;
  setAmountDrafts: React.Dispatch<React.SetStateAction<Record<string, string>>>;
  onCommitAmount: (row: CalculationRow) => Promise<void>;
  onAllocation: (id: string, allocation: Allocation) => Promise<void>;
  onToggle: (id: string) => Promise<void>;
  toggleLabel: string;
  editable: boolean;
};

function TransactionTable(props: TableProps) {
  return (
    <div className="table-block">
      <div className="section-head">
        <h3>{props.title}</h3>
        <span>{props.rows.length} {props.rows.length === 1 ? "transaction" : "transactions"}</span>
      </div>
      {props.rows.length === 0 ? (
        <p className="table-empty">No transactions in this group.</p>
      ) : (
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Date</th>
                <th>Description</th>
                <th className="number">Imported amount</th>
                {props.editable && <th>Amount to split</th>}
                {props.editable && <th>Allocation</th>}
                <th aria-label="Actions"></th>
              </tr>
            </thead>
            <tbody>
              {props.rows.map((row) => (
                <tr key={row.id}>
                  <td className="date date-cell">{row.date}</td>
                  <td className="description-cell"><strong>{row.description}</strong><small>{row.source_file}{row.source_line ? ` · line ${row.source_line}` : ""}</small></td>
                  <td className="number imported-cell">{formatCents(props.currency, row.amount_cents)}</td>
                  {props.editable && (
                    <td className="split-cell">
                      <input
                        className="amount"
                        aria-label={`Amount to split for ${row.description}`}
                        value={props.amountDrafts[row.id] ?? amountInput(row.split_amount_cents)}
                        onChange={(event) => props.setAmountDrafts((drafts) => ({ ...drafts, [row.id]: event.target.value }))}
                        onBlur={() => void props.onCommitAmount(row)}
                        onKeyDown={(event) => {
                          if (event.key === "Enter") event.currentTarget.blur();
                        }}
                      />
                    </td>
                  )}
                  {props.editable && (
                    <td className="allocation-cell">
                      <select aria-label={`Allocation for ${row.description}`} value={row.allocation} onChange={(event) => void props.onAllocation(row.id, event.target.value as Allocation)}>
                        <option value="split_evenly">Split evenly</option>
                        <option value="participant_one">Participant one</option>
                        <option value="participant_two">Participant two</option>
                      </select>
                    </td>
                  )}
                  <td className="action"><button className="text-button" aria-label={`${props.toggleLabel} ${row.description}`} onClick={() => void props.onToggle(row.id)}>{props.toggleLabel}</button></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

async function requestAPI<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, init);
  const payload = await response.json() as APIEnvelope<T>;
  if (!response.ok || payload.error) {
    throw new Error(payload.error?.message ?? `Request failed with status ${response.status}.`);
  }
  if (payload.data === undefined) throw new Error("The API returned an empty response.");
  return payload.data;
}

function formatCents(currency: string, cents: number): string {
  const sign = cents < 0 ? "−" : "";
  const absolute = Math.abs(cents);
  const whole = Math.floor(absolute / 100).toLocaleString("sv-SE");
  return `${sign}${currency.trim()} ${whole},${String(absolute % 100).padStart(2, "0")}`;
}

function formatHalfCents(currency: string, halfCents: number): string {
  const sign = halfCents < 0 ? "−" : "";
  const thousandths = Math.abs(halfCents) * 5;
  const whole = Math.floor(thousandths / 1000).toLocaleString("sv-SE");
  const remainder = thousandths % 1000;
  const fraction = remainder % 10 === 0
    ? String(remainder / 10).padStart(2, "0")
    : String(remainder).padStart(3, "0");
  return `${sign}${currency.trim()} ${whole},${fraction}`;
}

function amountInput(cents: number): string {
  const sign = cents < 0 ? "-" : "";
  const absolute = Math.abs(cents);
  return `${sign}${Math.floor(absolute / 100)},${String(absolute % 100).padStart(2, "0")}`;
}

function parseAmountCents(value: string): number | null {
  const normalized = value.trim().replace(/[\s\u00a0]/g, "").replace(",", ".");
  if (!/^[+-]?\d+(?:\.\d{1,2})?$/.test(normalized)) return null;
  const cents = Math.round(Number(normalized) * 100);
  return Number.isSafeInteger(cents) ? cents : null;
}

function messageFrom(reason: unknown): string {
  return reason instanceof Error ? reason.message : "Something went wrong while contacting the API.";
}
