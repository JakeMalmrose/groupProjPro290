using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;

public class Library
{
    [Key]
    public int ID { get; set; }  // Primary key

    // Navigation properties
    public virtual ICollection<User> Users { get; set; }
    public virtual ICollection<LibraryGame> LibraryGames { get; set; }
}
