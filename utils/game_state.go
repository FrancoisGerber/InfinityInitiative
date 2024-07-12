package utils

type GameState struct {
	ResourceListState bool `json:"ResourceListState" form:"ResourceListState" query:"ResourceListState"`
	InventoryState    bool `json:"InventoryState" form:"InventoryState" query:"InventoryState"`
	NotesState        bool `json:"NotesState" form:"NotesState" query:"NotesState"`
}
