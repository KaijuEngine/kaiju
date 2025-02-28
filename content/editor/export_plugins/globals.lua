---@class kmath
kmath = {}

---@class kstring
kstring = {}

---@shape ktable
ktable = {}

--- Returns the value clamped between the min and max (inclusive)
---@param current number
---@param min number
---@param max number
---@return number
function kmath.clamp(current, min, max)
	return math.max(min, math.min(max, current))
end

--- Get the count of all items in the table
---@param t table
---@return number
function ktable.len(t)
	local count = 0
	for _ in pairs(t) do count = count + 1 end
	return count
end

--- Find the index of a given element in the table
---@generic T
---@param t T[]
---@param elm T
---@return number
function ktable.index_of(t, elm)
	local idx = -1
	for i, v in ipairs(t) do
		if v == elm then
			idx = i
			break
		end
	end
	return idx
end

---remove_occurrence
---@generic T
---@param array T[]
---@param compare T
function ktable.remove_occurrence(array, compare)
	for i, instance in ipairs(array) do
		if instance == compare then
			table.remove(array, i)
			break
		end
	end
end

--- Clears out a table's contents
---@param t table
function ktable.clear(t)
	local count = #t
	for i=1, count do t[i]=nil end
end

--- Pretty prints a table
---@param t table
---@param tabs number
function ktable.print_table(t, tabs)
	tabs = tabs or 0
	local tab = "\t"
	for k, v in pairs(t) do
		if type(v) == "table" then
			print(tab:rep(tabs) .. tostring(k), "{}")
			ktable.print_table(v, tabs + 1)
		else
			print(tab:rep(tabs) .. tostring(k), v)
		end
	end
end

--- Find an entry in the table matching the query
---@param t table
---@param f function Predicate function For locating elements
---@return any|nil
function ktable.find(t, f)
	for k, v in pairs(t) do
		if f(v) then return v end
	end
	return nil
end

--- Find the index of the entry in the table matching the query
---@param t T[]
---@param f T|fun(entry:T):boolean Predicate function For locating elements
---@generic T
---@return number
function ktable.find_index(t, f)
	if type(f) == "function" then
		for i, v in ipairs(t) do
			if f(v) then return i end
		end
	else
		for i, v in ipairs(t) do
			if v == f then return i end
		end
	end
	return 0
end

---any
---@generic T
---@param t table<any,T>
---@param predicate fun(entry:T):boolean
---@return boolean
function ktable.any(t, predicate)
	for k, v in pairs(t) do
		if predicate(v) then
			return true
		end
	end
	return false
end

---select
---@generic T
---@generic V
---@param t table<any,T>
---@param predicate fun(entry:T):V
function ktable.select(t, predicate)
	---@type V[]
	local select = {}
	for _, v in pairs(t) do
		table.insert(select, predicate(v))
	end
	return select
end

---where
---@generic T
---@param t table<any,T>
---@param predicate fun(entry:T):boolean
---@return T[]
function ktable.where(t, predicate)
	---@type T[]
	local select = {}
	for _, v in pairs(t) do
		if predicate(v) then
			table.insert(select, v)
		end
	end
	return select
end

--- Goes through the given table and change it by removing any entries where
--- the predicate returns true
---@generic T
---@param t T[]
---@param predicate fun(entry:T):boolean
function ktable.remove_where(t, predicate)
	for i, v in ipairs(t) do
		if predicate(v) then
			t[i] = nil
		end
	end
	ktable.compact(t)
end

--- Search to see if a table has a given element in it
---@param t table
---@param search any
---@return boolean
function ktable.includes(t, search)
	for k, v in pairs(t) do
		if v == search then return true end
	end
	return false
end

---Joins multiple arrays together
---@return any[]
function ktable.join(...)
	local count = 0
	local args = { ... }
	for i=1, #args do
		count = count + #args[i]
	end
	local out = --[[---@type any[] ]] presized_table(count)
	local idx = 1
	for i=1, #args do
		for j=1, #args[i] do
			out[idx] = args[i][j]
			idx = idx + 1
		end
	end
	return out
end

--- Goes through the table and compares against nil, if found it removes it
--- (every instacne) from the array more efficiently than table.remove
---@param t table
function ktable.compact(t)
	local j = 1
	for i, v in pairs(t) do
		if t[i] ~= nil then
			if i ~= j then
				t[j] = v
				t[i] = nil
			end
			j = j + 1
		end
	end
end

--- Copy one table to another (including all members)
---@generic T
---@param t T
---@param seen table<T,T>
---@overload fun<T>(t:T):T
---@overload fun<T>(t:T,seen:table<T,T>):T
---@return T
function ktable.clone(t, seen)
	if type(t) ~= 'table' then return t end
	if seen and seen[t] then return seen[t] end
	local s = seen or {}
	---@type T
	local res = setmetatable({}, getmetatable(t))
	s[t] = res
	for k, v in pairs(t) do res[ktable.clone(k, s)] = ktable.clone(v, s) end
	return res
end

---copy
---@generic T
---@param t T
---@return T
function ktable.copy(t)
	---@type T
	local cpy = {}
	for k, v in pairs(t) do
		cpy[k] = v
	end
	return cpy
end

--- Reverse ipair
---@param t table
function ktable.ripairs(t)
	local i = #t
	return function()
		while i > 0 do
			local idx = i
			i = i - 1
			return idx, t[idx]
		end
	end
end

---move
---@param t T[]
---@param from number
---@param to number
---@generic T
function ktable.move(t, from, to)
	if from ~= to then
		local target = t[from]
		if to < from then
			from = from + 1
		else
			to = to + 1
		end
		table.insert(t, to, target)
		table.remove(t, from)
	end
end

---@shape WeightedTable
---@field public weight number

---weighted_select
---@generic T : WeightedTable
---@param t T[]
---@return T
function ktable.weighted_select(t)
	local sum = 0
	local selected
	local rand = math.random
	for i=1, #t do
		local entry = t[i]
		local r = rand(sum + entry.weight)
		if entry.weight >= r then
			selected = entry
		end
		sum = sum + entry.weight
	end
	return selected
end

---empty
---@param obj table|string
---@return boolean
function empty(obj)
	return obj == nil or #obj == 0
end

--- Check to see if a string starts with another string
---@param str string
---@param start string
---@return boolean
function kstring.starts_with(str, start)
	return str:sub(1, #start) == start
end

--- Check to see if a string ends with another string
---@param str string
---@param ending string
---@return boolean
function kstring.ends_with(str, ending)
	return ending == "" or str:sub(-#ending) == ending
end

--- Uppercase the first letter of a string
---@param str string
---@return string
function kstring.uc_first(str)
	local res, _ = str:gsub("^%l", string.upper)
	return res
end

--- Uppercase the first letter of each word
---@param str string
---@return string
function kstring.uc_words(str)
	local res, _ = str:gsub("%a", string.upper, 1)
	return res
end

---trim
---@param str string
---@return string
function kstring.trim(str)
	local res, _ = str:gsub("^%s*(.-)%s*$", "%1")
	return res
end

---split
---@param s string
---@param delimiter string
---@return string[]
function kstring.split(s, delimiter)
	---@type string[]
	local result = {}
	for match in (s..delimiter):gmatch("(.-)" .. delimiter) do
		result[#result+1] = --[[---@type string]] match
	end
	return result
end

--- Creates a class-like object
---@generic T
---@overload fun<T>(self:T):T
---@overload fun<T>(self:T,shallowCopy:boolean):T
---@param self T
---@param shallowCopy boolean
---@return T
function create_obj(self, shallowCopy)
	---@type T
	local o = --[[---@type T]] {}
	local mt = {}
	---@type table
	local base = self._extend
	---@type string[]
	local aka = {}
	if self.__name then
		table.insert(aka, self.__name)
	end
	---@type table[]
	local bases = {}
	while base ~= nil do
		table.insert(bases, base)
		if base.__name then
			table.insert(aka, base.__name)
		end
		base = base._extend
	end
	for i, b in ktable.ripairs(bases) do
		for k, v in pairs(b) do
			if type(v) == "function" and string.find(k, "^__") then
				mt[k] = v
			elseif type(v) == "table" then
				o[k] = ktable.clone(v)
			else
				o[k] = v
			end
		end
	end
	for k, v in pairs(self) do
		if type(v) == "function" and string.find(k, "^__") then
			mt[k] = v
		elseif type(v) == "table" then
			if shallowCopy then
				local tbl = {}
				for tk, tv in pairs(v) do
					tbl[tk] = tv
				end
				o[k] = tbl
			else
				o[k] = ktable.clone(v)
			end
		else
			o[k] = v
		end
	end
	setmetatable(o, mt)
	--o.__index = self
	o._extend = nil
	if #aka > 0 then
		-- TODO:  Make sure these are unique aka names
		o.__v_aka = aka
	end
	return o
end

--- Check if obj is or a child of is
---@generic T
---@param obj T
---@param typeName string
function instanceof(obj, typeName)
	if obj and obj.__v_aka then
		for i=1, #obj.__v_aka do
			if obj.__v_aka[i] == typeName then
				return true
			end
		end
	end
	return false
end

---vec2
---@param x number
---@param y number
---@return InlineVec2
function vec2(x, y) return { x = x, y = y } end

---vec3
---@param x number
---@param y number
---@param z number
---@return InlineVec3
function vec3(x, y, z) return { x = x, y = y, z = z } end

---vec4
---@param x number
---@param y number
---@param z number
---@param w number
---@return InlineVec4
function vec4(x, y, z, w) return { x = x, y = y, z = z, w = w } end

---color
---@param r number
---@param g number
---@param b number
---@param a number
---@return InlineColor
function color(r, g, b, a) return { r = r, g = g, b = b, a = a } end

---color_mix
---@param lhs InlineColor
---@param rhs InlineColor
---@param amount number
---@return InlineColor
function color_mix(lhs, rhs, amount)
	return color(lhs.r + (rhs.r - lhs.r) * amount,
		lhs.g + (rhs.g - lhs.g) * amount,
		lhs.b + (rhs.b - lhs.b) * amount,
		lhs.a + (rhs.a - lhs.a) * amount)
end

---@return InlineColor
function color_red() return color(1, 0, 0, 1) end
---@return InlineColor
function color_white() return color(1, 1, 1, 1) end
---@return InlineColor
function color_blue() return color(0, 0, 1, 1) end
---@return InlineColor
function color_black() return color(0, 0, 0, 1) end
---@return InlineColor
function color_green() return color(0, 1, 0, 1) end
---@return InlineColor
function color_yellow() return color(1, 1, 0, 1) end
---@return InlineColor
function color_orange() return color(1, 0.647, 0, 1) end
---@return InlineColor
function color_clear() return color(0, 0, 0, 0) end
---@return InlineColor
function color_gray() return color(0.5, 0.5, 0.5, 1) end
---@return InlineColor
function color_purple() return color(0.5, 0, 0.5, 1) end
---@return InlineColor
function color_brown() return color(0.647, 0.165, 0.165, 1) end
---@return InlineColor
function color_pink() return color(1, 0.753, 0.796, 1) end
---@return InlineColor
function color_cyan() return color(0, 1, 1, 1) end
---@return InlineColor
function color_magenta() return color(1, 0, 1, 1) end
---@return InlineColor
function color_teal() return color(0, 0.5, 0.5, 1) end
---@return InlineColor
function color_lime() return color(0, 1, 0, 1) end
---@return InlineColor
function color_maroon() return color(0.5, 0, 0, 1) end
---@return InlineColor
function color_olive() return color(0.5, 0.5, 0, 1) end
---@return InlineColor
function color_navy() return color(0, 0, 0.5, 1) end
---@return InlineColor
function color_silver() return color(0.753, 0.753, 0.753, 1) end
---@return InlineColor
function color_gold() return color(1, 0.843, 0, 1) end
---@return InlineColor
function color_sky() return color(0.529, 0.808, 0.922, 1) end
---@return InlineColor
function color_violet() return color(0.933, 0.51, 0.933, 1) end
---@return InlineColor
function color_indigo() return color(0.294, 0, 0.51, 1) end
---@return InlineColor
function color_turquoise() return color(0.251, 0.878, 0.816, 1) end
---@return InlineColor
function color_azure() return color(0.941, 1, 1, 1) end
---@return InlineColor
function color_chartreuse() return color(0.498, 1, 0, 1) end
---@return InlineColor
function color_coral() return color(1, 0.498, 0.314, 1) end
---@return InlineColor
function color_crimson() return color(0.863, 0.078, 0.235, 1) end
---@return InlineColor
function color_fuchsia() return color(1, 0, 1, 1) end
---@return InlineColor
function color_khaki() return color(0.941, 0.902, 0.549, 1) end
---@return InlineColor
function color_lavender() return color(0.902, 0.902, 0.98, 1) end
---@return InlineColor
function color_moccasin() return color(1, 0.894, 0.71, 1) end
---@return InlineColor
function color_salmon() return color(0.98, 0.502, 0.447, 1) end
---@return InlineColor
function color_sienna() return color(0.627, 0.322, 0.176, 1) end
---@return InlineColor
function color_tan() return color(0.824, 0.706, 0.549, 1) end
---@return InlineColor
function color_tomato() return color(1, 0.388, 0.278, 1) end
---@return InlineColor
function color_wheat() return color(0.961, 0.871, 0.702, 1) end
---@return InlineColor
function color_aqua() return color(0, 1, 1, 1) end
---@return InlineColor
function color_aquamarine() return color(0.498, 1, 0.831, 1) end
---@return InlineColor
function color_beige() return color(0.961, 0.961, 0.863, 1) end
---@return InlineColor
function color_bisque() return color(1, 0.894, 0.769, 1) end
---@return InlineColor
function color_blanchedalmond() return color(1, 0.922, 0.804, 1) end
---@return InlineColor
function color_blueviolet() return color(0.541, 0.169, 0.886, 1) end
---@return InlineColor
function color_burlywood() return color(0.871, 0.722, 0.529, 1) end
---@return InlineColor
function color_cadetblue() return color(0.373, 0.62, 0.627, 1) end
---@return InlineColor
function color_chocolate() return color(0.824, 0.412, 0.118, 1) end
---@return InlineColor
function color_cornflowerblue() return color(0.392, 0.584, 0.929, 1) end
---@return InlineColor
function color_cornsilk() return color(1, 0.973, 0.863, 1) end
---@return InlineColor
function color_darkblue() return color(0, 0, 0.545, 1) end
---@return InlineColor
function color_darkcyan() return color(0, 0.545, 0.545, 1) end
---@return InlineColor
function color_darkgoldenrod() return color(0.722, 0.525, 0.043, 1) end
---@return InlineColor
function color_darkgray() return color(0.663, 0.663, 0.663, 1) end
---@return InlineColor
function color_darkgreen() return color(0, 0.392, 0, 1) end
---@return InlineColor
function color_darkkhaki() return color(0.741, 0.718, 0.42, 1) end
---@return InlineColor
function color_darkmagenta() return color(0.545, 0, 0.545, 1) end
---@return InlineColor
function color_darkolivegreen() return color(0.333, 0.42, 0.184, 1) end
---@return InlineColor
function color_darkorange() return color(1, 0.549, 0, 1) end
---@return InlineColor
function color_darkorchid() return color(0.6, 0.196, 0.8, 1) end
---@return InlineColor
function color_darkred() return color(0.545, 0, 0, 1) end
---@return InlineColor
function color_darksalmon() return color(0.914, 0.588, 0.478, 1) end
---@return InlineColor
function color_darkseagreen() return color(0.561, 0.737, 0.561, 1) end
---@return InlineColor
function color_darkslateblue() return color(0.282, 0.239, 0.545, 1) end
---@return InlineColor
function color_darkslategray() return color(0.184, 0.31, 0.31, 1) end
---@return InlineColor
function color_darkturquoise() return color(0, 0.808, 0.82, 1) end
---@return InlineColor
function color_darkviolet() return color(0.58, 0, 0.827, 1) end
---@return InlineColor
function color_deeppink() return color(1, 0.078, 0.576, 1) end
---@return InlineColor
function color_deepskyblue() return color(0, 0.749, 1, 1) end
---@return InlineColor
function color_dimgray() return color(0.412, 0.412, 0.412, 1) end
---@return InlineColor
function color_dodgerblue() return color(0.118, 0.565, 1, 1) end
---@return InlineColor
function color_firebrick() return color(0.698, 0.133, 0.133, 1) end
---@return InlineColor
function color_floralwhite() return color(1, 0.98, 0.941, 1) end
---@return InlineColor
function color_forestgreen() return color(0.133, 0.545, 0.133, 1) end
---@return InlineColor
function color_gainsboro() return color(0.863, 0.863, 0.863, 1) end
---@return InlineColor
function color_ghostwhite() return color(0.973, 0.973, 1, 1) end
---@return InlineColor
function color_goldenrod() return color(0.855, 0.647, 0.125, 1) end
---@return InlineColor
function color_greenyellow() return color(0.678, 1, 0.184, 1) end
---@return InlineColor
function color_honeydew() return color(0.941, 1, 0.941, 1) end
---@return InlineColor
function color_hotpink() return color(1, 0.412, 0.706, 1) end
---@return InlineColor
function color_indianred() return color(0.804, 0.361, 0.361, 1) end
---@return InlineColor
function color_ivory() return color(1, 1, 0.941, 1) end
---@return InlineColor
function color_lavenderblush() return color(1, 0.941, 0.961, 1) end
---@return InlineColor
function color_lawngreen() return color(0.486, 0.988, 0, 1) end
---@return InlineColor
function color_lemonchiffon() return color(1, 0.98, 0.804, 1) end
---@return InlineColor
function color_lightblue() return color(0.678, 0.847, 0.902, 1) end
---@return InlineColor
function color_lightcoral() return color(0.941, 0.502, 0.502, 1) end
---@return InlineColor
function color_lightcyan() return color(0.878, 1, 1, 1) end
---@return InlineColor
function color_lightgoldenrodyellow() return color(0.98, 0.98, 0.824, 1) end
---@return InlineColor
function color_lightgreen() return color(0.565, 0.933, 0.565, 1) end
---@return InlineColor
function color_lightgrey() return color(0.827, 0.827, 0.827, 1) end
---@return InlineColor
function color_lightpink() return color(1, 0.714, 0.757, 1) end
---@return InlineColor
function color_lightsalmon() return color(1, 0.627, 0.478, 1) end
---@return InlineColor
function color_lightseagreen() return color(0.125, 0.698, 0.667, 1) end
---@return InlineColor
function color_lightskyblue() return color(0.529, 0.808, 0.98, 1) end
---@return InlineColor
function color_lightslategray() return color(0.467, 0.533, 0.6, 1) end
---@return InlineColor
function color_lightsteelblue() return color(0.69, 0.769, 0.871, 1) end
---@return InlineColor
function color_lightyellow() return color(1, 1, 0.878, 1) end
---@return InlineColor
function color_limegreen() return color(0.196, 0.804, 0.196, 1) end
---@return InlineColor
function color_linen() return color(0.98, 0.941, 0.902, 1) end
---@return InlineColor
function color_mediumaquamarine() return color(0.4, 0.804, 0.667, 1) end
---@return InlineColor
function color_mediumblue() return color(0, 0, 0.804, 1) end
---@return InlineColor
function color_mediumorchid() return color(0.729, 0.333, 0.827, 1) end
---@return InlineColor
function color_mediumpurple() return color(0.576, 0.439, 0.859, 1) end
---@return InlineColor
function color_mediumseagreen() return color(0.235, 0.702, 0.443, 1) end
---@return InlineColor
function color_mediumslateblue() return color(0.482, 0.408, 0.933, 1) end
---@return InlineColor
function color_mediumspringgreen() return color(0, 0.98, 0.604, 1) end
---@return InlineColor
function color_mediumturquoise() return color(0.282, 0.82, 0.8, 1) end
---@return InlineColor
function color_mediumvioletred() return color(0.78, 0.082, 0.522, 1) end
---@return InlineColor
function color_midnightblue() return color(0.098, 0.098, 0.439, 1) end
---@return InlineColor
function color_mintcream() return color(0.961, 1, 0.98, 1) end
---@return InlineColor
function color_mistyrose() return color(1, 0.894, 0.882, 1) end
---@return InlineColor
function color_navajowhite() return color(1, 0.871, 0.678, 1) end
---@return InlineColor
function color_oldlace() return color(0.992, 0.961, 0.902, 1) end
---@return InlineColor
function color_olivedrab() return color(0.42, 0.557, 0.137, 1) end
---@return InlineColor
function color_orangered() return color(1, 0.271, 0, 1) end
---@return InlineColor
function color_orchid() return color(0.855, 0.439, 0.839, 1) end
---@return InlineColor
function color_palegoldenrod() return color(0.933, 0.91, 0.667, 1) end
---@return InlineColor
function color_palegreen() return color(0.596, 0.984, 0.596, 1) end
---@return InlineColor
function color_paleturquoise() return color(0.686, 0.933, 0.933, 1) end
---@return InlineColor
function color_palevioletred() return color(0.859, 0.439, 0.576, 1) end
---@return InlineColor
function color_papayawhip() return color(1, 0.937, 0.835, 1) end
---@return InlineColor
function color_peachpuff() return color(1, 0.855, 0.725, 1) end
---@return InlineColor
function color_peru() return color(0.804, 0.522, 0.247, 1) end
---@return InlineColor
function color_plum() return color(0.867, 0.627, 0.867, 1) end
---@return InlineColor
function color_powderblue() return color(0.69, 0.878, 0.902, 1) end
---@return InlineColor
function color_rosybrown() return color(0.737, 0.561, 0.561, 1) end
---@return InlineColor
function color_royalblue() return color(0.255, 0.412, 0.882, 1) end
---@return InlineColor
function color_saddlebrown() return color(0.545, 0.271, 0.075, 1) end
---@return InlineColor
function color_sandybrown() return color(0.957, 0.643, 0.376, 1) end
---@return InlineColor
function color_seagreen() return color(0.18, 0.545, 0.341, 1) end
---@return InlineColor
function color_seashell() return color(1, 0.961, 0.933, 1) end
---@return InlineColor
function color_skyblue() return color(0.529, 0.808, 0.922, 1) end
---@return InlineColor
function color_slateblue() return color(0.416, 0.353, 0.804, 1) end
---@return InlineColor
function color_slategray() return color(0.439, 0.502, 0.565, 1) end
---@return InlineColor
function color_slategrey() return color(0.439, 0.502, 0.565, 1) end
---@return InlineColor
function color_snow() return color(1, 0.98, 0.98, 1) end
---@return InlineColor
function color_springgreen() return color(0, 1, 0.498, 1) end
---@return InlineColor
function color_steelblue() return color(0.275, 0.51, 0.706, 1) end
---@return InlineColor
function color_thistle() return color(0.847, 0.749, 0.847, 1) end
---@return InlineColor
function color_whitesmoke() return color(0.961, 0.961, 0.961, 1) end
---@return InlineColor
function color_yellowgreen() return color(0.604, 0.804, 0.196, 1) end

---tag_factory_shader
---@param gameHost GameHost
---@param entity Entity
---@param args string
function tag_factory_shader(gameHost, entity, args)
	---@type string[]
	local parts = {}
	for token in string.gmatch(args, "[^,]+") do
		parts[#parts+1] = --[[---@type string]] token
	end
	local vert = parts[1] and "shaders/"..parts[1]..".vert" or SHADER_VERT_BASIC
	local frag = parts[2] and "shaders/"..parts[2]..".frag" or SHADER_FRAG_BASIC
	local geom = parts[3] and "shaders/"..parts[3]..".geom" or nil
	local shader = host_shader(gameHost.host, vert, frag, geom)
	entity:set_visual_shader(gameHost, shader)
end

---@shape Globals
---@field public host Host
---@field tags table
---@field destroyList Entity[]
Globals = {
	host = nil,
	Tags = {
		shader = tag_factory_shader
	},
	destroyList = {},
}

-- TODO:  This probably isn't needed since we are no longer on Lua 5.1
---Gives the table a destructor by creating proxy userdata and giving it a __gc.
---The function provided will be the function that is called upon destruction
---@param self table
---@param f function
function destructor(self, f)
	self._destructor = newproxy(true)
	getmetatable(self._destructor).__gc = f
end

--- Throws a failed assert with not implemented message
function not_implemented_error()
	print(debug.traceback())
	print(debug.getinfo(1))
	error("Not implemented", 0)
end

---@shape AsyncResolve
---@overload fun():void
---@overload fun(result:any):void
---
---@shape AsyncReject
---@overload fun():void
---@overload fun(error:any):void

---@alias AwaitFunction fun(await:Await)|fun(await:Await,_1:any)|fun(await:Await,_1:any,_2:any)|fun(await:Await,_1:any,_2:any,_3:any)|fun(await:Await,_1:any,_2:any,_3:any,_4:any)|fun(await:Await,_1:any,_2:any,_3:any,_4:any,_5:any)|fun(await:Await,_1:any,_2:any,_3:any,_4:any,_5:any,_6:any)|fun(await:Await,_1:any,_2:any,_3:any,_4:any,_5:any,_6:any,_7:any)|fun(await:Await,_1:any,_2:any,_3:any,_4:any,_5:any,_6:any,_7:any,_8:any)|fun(await:Await,_1:any,_2:any,_3:any,_4:any,_5:any,_6:any,_7:any,_8:any,_9:any)

---@class Await
---@field resolve AsyncResolve
---@field reject AsyncReject
---@field wait fun():any
---@overload fun(waitFunction:AwaitFunction,...:any):any

---await_resolution
---@param await Await
function await_resolve(await) end

---async
---@vararg any
---@param routine fun(await:Await)|fun(await:Await,_1:any)|fun(await:Await,_1:any,_2:any)|fun(await:Await,_1:any,_2:any,_3:any)|fun(await:Await,_1:any,_2:any,_3:any,_4:any)|fun(await:Await,_1:any,_2:any,_3:any,_4:any,_5:any)|fun(await:Await,_1:any,_2:any,_3:any,_4:any,_5:any,_6:any)|fun(await:Await,_1:any,_2:any,_3:any,_4:any,_5:any,_6:any,_7:any)|fun(await:Await,_1:any,_2:any,_3:any,_4:any,_5:any,_6:any,_7:any,_8:any)|fun(await:Await,_1:any,_2:any,_3:any,_4:any,_5:any,_6:any,_7:any,_8:any,_9:any)
function async(routine, ...)
	local f = coroutine.create(function(await, ...) routine(await, ...) end)
	local await = { error = nil, result = nil, completed = false }
	local complete = function(arg, err)
		await.result = arg
		await.error = err
		await.completed = true
		coroutine.resume(f)
	end
	await.resolve = function(arg) complete(arg, nil) end
	await.reject = function(err) complete(nil, err) end
	await.__call = function(self, wait, ...)
		local lastResult = self.result
		self.completed = false
		wait(self, ...)
		if not self.completed then coroutine.yield(f, ...) end
		if self.error then assert(false, self.error) end
		self.completed = false
		local newResult = self.result
		self.result = lastResult
		return newResult
	end
	await.wait = function() return await(await_resolve) end
	setmetatable(await, await)
	coroutine.resume(f, await, ...)
end

---valid_ip_address
---@param addr string
---@return boolean
function valid_ip_address(addr)
	-- TODO:  Match IPv6 as well
	local isValid = addr:match("%d+%.%d+%.%d+%.%d+") ~= nil
	isValid = isValid or (addr:match("[%a%d][%a%d][%a%d][%a%d]") ~= nil)
	if isValid then
		local parts = kstring.split(addr, ".")
		for i=1, #parts do
			if #parts[i] > 3 then
				isValid = false
			end
		end
	end
	return isValid
end

function memorize(f)
	local mem = {}									-- Memorizing table
	setmetatable(mem, {__mode = "kv"})	-- Make it weak
	return function (x)								-- New version of 'f', with memorizing
		local r = mem[x]
		if r == nil then							-- No previous result?
			r = f(x)								-- Calls original function
			mem[x] = r								-- Store result for reuse
		end
		return r
	end
end
