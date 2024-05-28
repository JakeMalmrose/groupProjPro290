using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;

public class User
{
    [Key]
    public int ID { get; set; }  // Primary key

    [Required]
    [MaxLength(50)]
    public string Username { get; set; }

    [Required]
    [MaxLength(255)]
    public string Password { get; set; }  // Assuming passwords are hashed

    [Required]
    public DateTime CreatedDate { get; set; }

    public int? CartID { get; set; }
    public int? LibraryID { get; set; }
    public float Balance { get; set; }

    // Navigation properties
    public virtual Cart Cart { get; set; }
    public virtual Library Library { get; set; }
    public virtual ICollection<UserRole> UserRoles { get; set; }
    public virtual ICollection<Order> Orders { get; set; }
}
