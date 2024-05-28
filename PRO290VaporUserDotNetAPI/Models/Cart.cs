using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;

public class Cart
{
    [Key]
    public int ID { get; set; }  // Primary key

    // Navigation properties
    public virtual ICollection<User> Users { get; set; }
}
