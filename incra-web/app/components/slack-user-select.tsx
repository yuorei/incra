import { useState, useRef, useEffect } from "react";

export type SlackUser = {
  id: string;
  name: string;
  real_name: string;
  display_name: string;
  profile_image: string;
  team_id?: string;
};

type Props = {
  users: SlackUser[];
  defaultUserId?: string;
  excludeUserId?: string;
  onSelect: (user: SlackUser | null) => void;
};

type MultiProps = {
  users: SlackUser[];
  excludeUserId?: string;
  onSelect: (users: SlackUser[]) => void;
};

export function SlackUserCell({ userId, users }: { userId?: string; users: SlackUser[] }) {
  if (!userId) return <span className="text-gray-400">-</span>;
  const user = users.find((u) => u.id === userId);
  if (!user) return <span className="font-mono text-xs text-gray-500">{userId}</span>;
  const displayName = user.display_name || user.real_name || user.name;
  const href = user.team_id
    ? `slack://user?team=${user.team_id}&id=${user.id}`
    : undefined;
  const inner = (
    <>
      <img src={user.profile_image} alt="" className="w-7 h-7 rounded-full shrink-0" />
      <div className="min-w-0">
        <div className="text-sm text-gray-900 dark:text-gray-100 font-medium truncate">{displayName}</div>
        {user.real_name && user.display_name && user.real_name !== user.display_name && (
          <div className="text-xs text-gray-500 dark:text-gray-400 truncate">{user.real_name}</div>
        )}
      </div>
    </>
  );
  if (href) {
    return (
      <a href={href} className="flex items-center gap-2 hover:opacity-80">
        {inner}
      </a>
    );
  }
  return <div className="flex items-center gap-2">{inner}</div>;
}

export function SlackUserSelect({ users, defaultUserId, excludeUserId, onSelect }: Props) {
  const defaultUser = defaultUserId
    ? users.find((u) => u.id === defaultUserId) ?? null
    : null;
  const [selected, setSelected] = useState<SlackUser | null>(defaultUser);
  const [query, setQuery] = useState("");
  const [isOpen, setIsOpen] = useState(false);
  const wrapperRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (wrapperRef.current && !wrapperRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const filtered = users.filter((u) => {
    if (excludeUserId && u.id === excludeUserId) return false;
    const q = query.toLowerCase();
    return (
      u.name.toLowerCase().includes(q) ||
      u.real_name.toLowerCase().includes(q) ||
      u.display_name.toLowerCase().includes(q)
    );
  });

  function handleSelect(user: SlackUser) {
    setSelected(user);
    setQuery("");
    setIsOpen(false);
    onSelect(user);
  }

  function handleClear() {
    setSelected(null);
    setQuery("");
    onSelect(null);
  }

  function displayLabel(u: SlackUser) {
    return u.display_name || u.real_name || u.name;
  }

  if (selected) {
    return (
      <div className="flex items-center gap-2 w-full border border-gray-300 dark:border-gray-600 dark:bg-gray-700 rounded px-3 py-2 text-sm">
        <img
          src={selected.profile_image}
          alt=""
          className="w-6 h-6 rounded-full shrink-0"
        />
        <span className="text-gray-900 dark:text-white font-medium truncate">
          {displayLabel(selected)}
        </span>
        {selected.real_name && selected.display_name && selected.real_name !== selected.display_name && (
          <span className="text-gray-500 dark:text-gray-400 truncate">
            ({selected.real_name})
          </span>
        )}
        <span className="text-gray-400 dark:text-gray-500 truncate">
          @{selected.name}
        </span>
        <button
          type="button"
          onClick={handleClear}
          className="ml-auto text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 shrink-0"
        >
          &times;
        </button>
      </div>
    );
  }

  return (
    <div ref={wrapperRef} className="relative">
      <input
        type="text"
        value={query}
        onChange={(e) => {
          setQuery(e.target.value);
          setIsOpen(true);
        }}
        onFocus={() => setIsOpen(true)}
        placeholder="ユーザーを検索..."
        className="w-full border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
      />
      {isOpen && (
        <ul className="absolute z-10 mt-1 w-full max-h-60 overflow-auto bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded shadow-lg">
          {filtered.length === 0 ? (
            <li className="px-3 py-2 text-sm text-gray-500 dark:text-gray-400">
              該当するユーザーがいません
            </li>
          ) : (
            filtered.map((u) => (
              <li key={u.id}>
                <button
                  type="button"
                  onClick={() => handleSelect(u)}
                  className="w-full flex items-center gap-2 px-3 py-2 text-sm hover:bg-blue-50 dark:hover:bg-gray-700 text-left"
                >
                  <img
                    src={u.profile_image}
                    alt=""
                    className="w-7 h-7 rounded-full shrink-0"
                  />
                  <div className="min-w-0">
                    <div className="flex items-center gap-1">
                      <span className="text-gray-900 dark:text-white font-medium truncate">
                        {displayLabel(u)}
                      </span>
                      {u.real_name && u.display_name && u.real_name !== u.display_name && (
                        <span className="text-gray-500 dark:text-gray-400 text-xs truncate">
                          ({u.real_name})
                        </span>
                      )}
                    </div>
                    <div className="text-gray-400 dark:text-gray-500 text-xs">
                      @{u.name}
                    </div>
                  </div>
                </button>
              </li>
            ))
          )}
        </ul>
      )}
    </div>
  );
}

export function SlackUserMultiSelect({ users, excludeUserId, onSelect }: MultiProps) {
  const [selected, setSelected] = useState<SlackUser[]>([]);
  const [query, setQuery] = useState("");
  const [isOpen, setIsOpen] = useState(false);
  const wrapperRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (wrapperRef.current && !wrapperRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const filtered = users.filter((u) => {
    if (excludeUserId && u.id === excludeUserId) return false;
    if (selected.some((s) => s.id === u.id)) return false;
    const q = query.toLowerCase();
    return (
      u.name.toLowerCase().includes(q) ||
      u.real_name.toLowerCase().includes(q) ||
      u.display_name.toLowerCase().includes(q)
    );
  });

  function handleSelect(user: SlackUser) {
    const next = [...selected, user];
    setSelected(next);
    setQuery("");
    setIsOpen(false);
    onSelect(next);
  }

  function handleRemove(userId: string) {
    const next = selected.filter((u) => u.id !== userId);
    setSelected(next);
    onSelect(next);
  }

  function displayLabel(u: SlackUser) {
    return u.display_name || u.real_name || u.name;
  }

  return (
    <div ref={wrapperRef} className="relative">
      <div className="min-h-[38px] w-full border border-gray-300 dark:border-gray-600 dark:bg-gray-700 rounded px-2 py-1 flex flex-wrap gap-1 cursor-text"
        onClick={() => setIsOpen(true)}
      >
        {selected.map((u) => (
          <span
            key={u.id}
            className="inline-flex items-center gap-1 bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200 rounded px-2 py-0.5 text-sm"
          >
            <img src={u.profile_image} alt="" className="w-4 h-4 rounded-full shrink-0" />
            {displayLabel(u)}
            <button
              type="button"
              onClick={(e) => { e.stopPropagation(); handleRemove(u.id); }}
              className="text-blue-500 dark:text-blue-300 hover:text-blue-700 dark:hover:text-blue-100 ml-0.5"
            >
              &times;
            </button>
          </span>
        ))}
        <input
          type="text"
          value={query}
          onChange={(e) => {
            setQuery(e.target.value);
            setIsOpen(true);
          }}
          onFocus={() => setIsOpen(true)}
          placeholder={selected.length === 0 ? "ユーザーを検索..." : "追加..."}
          aria-label="請求先ユーザーを検索"
          className="flex-1 min-w-[120px] bg-transparent dark:text-white text-sm focus:outline-none py-0.5"
        />
      </div>
      {isOpen && (
        <ul className="absolute z-10 mt-1 w-full max-h-60 overflow-auto bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded shadow-lg">
          {filtered.length === 0 ? (
            <li className="px-3 py-2 text-sm text-gray-500 dark:text-gray-400">
              該当するユーザーがいません
            </li>
          ) : (
            filtered.map((u) => (
              <li key={u.id}>
                <button
                  type="button"
                  onClick={() => handleSelect(u)}
                  className="w-full flex items-center gap-2 px-3 py-2 text-sm hover:bg-blue-50 dark:hover:bg-gray-700 text-left"
                >
                  <img
                    src={u.profile_image}
                    alt=""
                    className="w-7 h-7 rounded-full shrink-0"
                  />
                  <div className="min-w-0">
                    <div className="flex items-center gap-1">
                      <span className="text-gray-900 dark:text-white font-medium truncate">
                        {displayLabel(u)}
                      </span>
                      {u.real_name && u.display_name && u.real_name !== u.display_name && (
                        <span className="text-gray-500 dark:text-gray-400 text-xs truncate">
                          ({u.real_name})
                        </span>
                      )}
                    </div>
                    <div className="text-gray-400 dark:text-gray-500 text-xs">
                      @{u.name}
                    </div>
                  </div>
                </button>
              </li>
            ))
          )}
        </ul>
      )}
    </div>
  );
}
