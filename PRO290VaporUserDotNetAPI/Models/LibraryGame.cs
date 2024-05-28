   public class LibraryGame
    {
        public int LibraryID { get; set; }
        public int GameID { get; set; }

        // Navigation properties
        public virtual Library Library { get; set; }
        public virtual Game Game { get; set; }
    }