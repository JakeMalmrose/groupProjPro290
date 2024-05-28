  public class UserRole
    {
        public int UserID { get; set; }
        public int RoleID { get; set; }

        // Navigation properties
        public virtual User User { get; set; }
        public virtual Role Role { get; set; }
    }