import { useState, useRef, useEffect } from "react";

export type SlackUser = {
  id: string;
  name: string;
  real_name: string;
  display_name: string;
  profile_image: string;
};

type Props = {
  users: SlackUser[];
  defaultUserId?: string;
  excludeUserId?: string;
  onSelect: (user: SlackUser | null) => void;
};

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
