package database

func (db *appdb) AddToGroup(groupID string, userIDToAdd string) error {
	_, err := db.c.Exec("INSERT INTO conversation_members (conversation_id, user_id) VALUES (?, ?)", groupID, userIDToAdd)
	return err
}

func (db *appdb) LeaveGroup(groupID string, userID string) error {
	_, err := db.c.Exec("DELETE FROM conversation_members WHERE conversation_id = ? AND user_id = ?", groupID, userID)
	return err
}

func (db *appdb) SetGroupName(groupID string, newName string) error {
	_, err := db.c.Exec("UPDATE conversations SET name = ? WHERE id = ? AND is_group = 1", newName, groupID)
	return err
}

func (db *appdb) SetGroupPhoto(groupID string, photoURL string) error {
	_, err := db.c.Exec("UPDATE conversations SET photo_url = ? WHERE id = ? AND is_group = 1", photoURL, groupID)
	return err
}
