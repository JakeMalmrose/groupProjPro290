using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;

public class Role
{
    [Key]
    public int ID { get; set; }  // Primary key

    [Required]
    [MaxLength(50)]
    public string RoleName { get; set; }

    public virtual ICollection<UserRole> UserRoles { get; set; }
}
